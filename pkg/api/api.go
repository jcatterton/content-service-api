package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"content-service-api/models"
	"content-service-api/pkg/dao"
	"content-service-api/pkg/external"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ListenAndServe() error {
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Access-Control-Allow-Origin", "Content-Type"})
	origins := handlers.AllowedOrigins([]string{"*"})
	methods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

	router, err := route()
	if err != nil {
		return err
	}

	server := &http.Server{
		Handler:      handlers.CORS(headers, origins, methods)(router),
		Addr:         ":8005",
		WriteTimeout: 20 * time.Second,
		ReadTimeout:  20 * time.Second,
	}
	shutdownGracefully(server)

	logrus.Info("Starting API server...")
	return server.ListenAndServe()
}

func route() (*mux.Router, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		logrus.WithError(err).Error("Error creating mongo client")
		return nil, err
	}

	dbHandler := dao.Handler{
		Client:          client,
		Database:        os.Getenv("DATABASE"),
		FileCollection:  os.Getenv("FILE_COLLECTION"),
		FsCollection:    os.Getenv("FS_COLLECTION"),
		ChunkCollection: os.Getenv("CHUNK_COLLECTION"),
	}

	extHandler := external.Handler{
		HttpClient:      &http.Client{Timeout: 5 * time.Second},
		LoginServiceURL: os.Getenv("LOGIN_SERVICE_URL"),
	}

	r := mux.NewRouter()

	r.HandleFunc("/health", checkHealth(&dbHandler)).Methods(http.MethodGet)
	r.HandleFunc("/upload", uploadFile(&dbHandler, &extHandler)).Methods(http.MethodPost)
	r.HandleFunc("/file/{id}", downloadFile(&dbHandler, &extHandler)).Methods(http.MethodGet)
	r.HandleFunc("/file/{id}", deleteFile(&dbHandler, &extHandler)).Methods(http.MethodDelete)
	r.HandleFunc("/file/{id}", updateFileInfo(&dbHandler, &extHandler)).Methods(http.MethodPut)
	r.HandleFunc("/files", getFiles(&dbHandler, &extHandler)).Methods(http.MethodGet)

	return r, nil
}

func checkHealth(handler dao.DBHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer closeRequestBody(r)
		if err := handler.Ping(r.Context()); err != nil {
			respondWithError(w, http.StatusInternalServerError, "API is running but unable to connect to database")
			return
		}
		respondWithSuccess(w, http.StatusOK, "API is running and connected to database")
		return
	}
}

func uploadFile(dbHandler dao.DBHandler, extHandler external.ExtHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer closeRequestBody(r)

		token, err := getAuthToken(r)
		if err != nil {
			logrus.WithError(err).Error("Error retrieving authorization token from request")
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := extHandler.ValidateToken(token); err != nil {
			logrus.WithError(err).Error("Error validating token")
			respondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}

		if err := r.ParseForm(); err != nil {
			logrus.WithError(err).Error("Error parsing request form")
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			logrus.WithError(err).Error("Error getting file from request")
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		defer func() {
			if err := file.Close(); err != nil {
				logrus.WithError(err).Error("Error closing file")
			}
		}()

		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, file); err != nil {
			logrus.WithError(err).Error("Error reading file")
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		uploadRequest := models.FileRequest{
			Name:      header.Filename,
			Timestamp: time.Now(),
			Extension: filepath.Ext(header.Filename),
			Size:      header.Size,
		}

		if err := dbHandler.UploadFile(ctx, &uploadRequest, buf.Bytes()); err != nil {
			logrus.WithError(err).Error("Error uploading file")
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		logrus.Info("File uploaded successfully")
		respondWithSuccess(w, http.StatusOK, "File uploaded successfully")
		return
	}
}

func downloadFile(dbHandler dao.DBHandler, extHandler external.ExtHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer closeRequestBody(r)

		token, err := getAuthToken(r)
		if err != nil {
			logrus.WithError(err).Error("Error retrieving authorization token from request")
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := extHandler.ValidateToken(token); err != nil {
			logrus.WithError(err).Error("Error validating token")
			respondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}

		id, err := primitive.ObjectIDFromHex(mux.Vars(r)["id"])
		if err != nil {
			logrus.WithError(err).Error("Error converting ID to ObjectID")
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		fileBytes, err := dbHandler.GetFile(ctx, id)
		if err != nil {
			logrus.WithError(err).Error("Error downloading file")
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if _, err := io.Copy(w, bytes.NewBuffer(fileBytes)); err != nil {
			logrus.WithError(err).Error("Error writing file to response")
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		logrus.Info("File successfully retrieved")
		return
	}
}

func deleteFile(dbHandler dao.DBHandler, extHandler external.ExtHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer closeRequestBody(r)

		token, err := getAuthToken(r)
		if err != nil {
			logrus.WithError(err).Error("Error retrieving authorization token from request")
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := extHandler.ValidateToken(token); err != nil {
			logrus.WithError(err).Error("Error validating token")
			respondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}

		id, err := primitive.ObjectIDFromHex(mux.Vars(r)["id"])
		if err != nil {
			logrus.WithError(err).Error("Error converting ID to ObjectID")
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := dbHandler.DeleteFile(ctx, id); err != nil {
			logrus.WithError(err).Error("Error downloading file")
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		logrus.Info("File successfully deleted")
		respondWithSuccess(w, http.StatusOK, "File successfully deleted")
		return
	}
}

func updateFileInfo(dbHandler dao.DBHandler, extHandler external.ExtHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer closeRequestBody(r)

		token, err := getAuthToken(r)
		if err != nil {
			logrus.WithError(err).Error("Error retrieving authorization token from request")
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := extHandler.ValidateToken(token); err != nil {
			logrus.WithError(err).Error("Error validating token")
			respondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}

		id, err := primitive.ObjectIDFromHex(mux.Vars(r)["id"])
		if err != nil {
			logrus.WithError(err).Error("Error converting ID to ObjectID")
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		var updateRequest map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
			logrus.WithError(err).Error("Error decoding request body")
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := dbHandler.UpdateFileInfo(ctx, id, updateRequest); err != nil {
			logrus.WithError(err).Error("Error updating file info")
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		logrus.Info("File updated successfully")
		respondWithSuccess(w, http.StatusOK, "File updated successfully")
		return
	}
}

func getFiles(dbHandler dao.DBHandler, extHandler external.ExtHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer closeRequestBody(r)

		token, err := getAuthToken(r)
		if err != nil {
			logrus.WithError(err).Error("Error retrieving authorization token from request")
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := extHandler.ValidateToken(token); err != nil {
			logrus.WithError(err).Error("Error validating token")
			respondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}

		query := make(map[string]interface{})
		for key, val := range r.URL.Query() {
			if key == "size" {
				v, err := strconv.Atoi(val[0])
				if err != nil {
					logrus.WithError(err).Warn("Error converting 'size' query parameter to int, skipping this parameter")
					continue
				}
				query[key] = v
				continue
			}
			query[key] = val[0]
		}

		results, err := dbHandler.GetFiles(ctx, query)
		if err != nil {
			logrus.WithError(err).Error("Error retrieving files from database")
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		logrus.Info("Files retrieved successfully")
		respondWithSuccess(w, http.StatusOK, results)
		return
	}
}

func shutdownGracefully(server *http.Server) {
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)
		<-signals

		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(c); err != nil {
			logrus.WithError(err).Error("Error shutting down server")
		}

		<-c.Done()
		os.Exit(0)
	}()
}

func respondWithSuccess(w http.ResponseWriter, code int, body interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if body == nil {
		logrus.Error("Body is nil, unable to write response")
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		logrus.WithError(err).Error("Error encoding response")
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if message == "" {
		logrus.Error("Body is nil, unable to write response")
		return
	}
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		logrus.WithError(err).Error("Error encoding response")
	}
}

func closeRequestBody(req *http.Request) {
	if req.Body == nil {
		return
	}
	if err := req.Body.Close(); err != nil {
		logrus.WithError(err).Error("Error closing request body")
		return
	}
	return
}

func getAuthToken(r *http.Request) (string, error) {
	tokenHeader := r.Header.Get("Authorization")
	if tokenHeader == "" {
		return "", errors.New("no authorization header found")
	} else if (len(tokenHeader) >= 7 && tokenHeader[:7] != "Bearer ") || len(strings.Split(tokenHeader, " ")) != 2 {
		return "", errors.New("authorization header must be in format 'Bearer' <token>")
	}
	return strings.Split(tokenHeader, " ")[1], nil
}
