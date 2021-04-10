package api

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"content-service-api/models"
	"content-service-api/pkg/testhelper/mocks"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestApi_CheckHealth_ShouldReturn500IfUnableToConnectToDatabase(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	dbHandler.On("Ping", mock.Anything).Return(errors.New("test"))

	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	require.Nil(t, err)

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(checkHealth(dbHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestApi_CheckHealth_ShouldReturn200OnSuccess(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	dbHandler.On("Ping", mock.Anything).Return(nil)

	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	require.Nil(t, err)

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(checkHealth(dbHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestApi_UploadFile_ShouldReturn400OnNoAuthorizationTokenFound(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}

	req, err := http.NewRequest(http.MethodPost, "/upload", nil)
	require.Nil(t, err)

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(uploadFile(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestApi_UploadFile_ShouldReturn401IfErrorOccursValidatingToken(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	extHandler.On("ValidateToken", mock.Anything).Return(errors.New("test"))

	req, err := http.NewRequest(http.MethodPost, "/upload", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(uploadFile(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestApi_UploadFile_ShouldReturn400IfErrorOccursParsingForm(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	extHandler.On("ValidateToken", mock.Anything).Return(nil)

	req, err := http.NewRequest(http.MethodPost, "/upload", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(uploadFile(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestApi_UploadFile_ShouldReturn400IfNoFormFieldWithKeyFileFound(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	extHandler.On("ValidateToken", mock.Anything).Return(nil)

	req, err := http.NewRequest(http.MethodPost, "/upload", strings.NewReader("{}"))
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(uploadFile(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestApi_UploadFile_ShouldReturn500OnDbHandlerError(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	dbHandler.On("UploadFile", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("test"))
	extHandler.On("ValidateToken", mock.Anything).Return(nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.png")
	require.Nil(t, err)

	_, err = io.Copy(part, bytes.NewBuffer([]byte("test")))
	require.Nil(t, err)

	require.Nil(t, writer.Close())

	req, err := http.NewRequest(http.MethodPost, "/upload", body)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req.Header.Add("Content-Type", writer.FormDataContentType())

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(uploadFile(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestApi_UploadFile_ShouldReturn200OnSuccess(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	dbHandler.On("UploadFile", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	extHandler.On("ValidateToken", mock.Anything).Return(nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.png")
	require.Nil(t, err)

	_, err = io.Copy(part, bytes.NewBuffer([]byte("test")))
	require.Nil(t, err)

	require.Nil(t, writer.Close())

	req, err := http.NewRequest(http.MethodPost, "/upload", body)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req.Header.Add("Content-Type", writer.FormDataContentType())

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(uploadFile(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestApi_DownloadFile_ShouldReturn400OnNoAuthorizationTokenFound(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}

	req, err := http.NewRequest(http.MethodGet, "/file/5df25cc42d811e3b6b945c08", nil)
	require.Nil(t, err)

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(downloadFile(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestApi_DownloadFile_ShouldReturn401IfErrorOccursValidatingToken(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	extHandler.On("ValidateToken", mock.Anything).Return(errors.New("test"))

	req, err := http.NewRequest(http.MethodGet, "/file/5df25cc42d811e3b6b945c08", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(downloadFile(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestApi_DownloadFile_ShouldReturn400IfUnableToCreateObjectIDFromGivenIDVar(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	extHandler.On("ValidateToken", mock.Anything).Return(nil)

	req, err := http.NewRequest(http.MethodGet, "/file/5df25cc42d811e3b6b945c08", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(downloadFile(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestApi_DownloadFile_ShouldReturn500OnHandlerError(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	dbHandler.On("GetFile", mock.Anything, mock.Anything).Return(nil, errors.New("test"))
	extHandler.On("ValidateToken", mock.Anything).Return(nil)

	req, err := http.NewRequest(http.MethodGet, "/file/5df25cc42d811e3b6b945c08", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req = mux.SetURLVars(req, map[string]string{"id": "5df25cc42d811e3b6b945c08"})

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(downloadFile(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestApi_DownloadFile_ShouldReturn200OnHandlerError(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	dbHandler.On("GetFile", mock.Anything, mock.Anything).Return([]byte{}, nil)
	extHandler.On("ValidateToken", mock.Anything).Return(nil)

	req, err := http.NewRequest(http.MethodGet, "/file/5df25cc42d811e3b6b945c08", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req = mux.SetURLVars(req, map[string]string{"id": "5df25cc42d811e3b6b945c08"})

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(downloadFile(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestApi_DeleteFile_ShouldReturn400OnNoAuthorizationTokenFound(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}

	req, err := http.NewRequest(http.MethodDelete, "/file/5df25cc42d811e3b6b945c08", nil)
	require.Nil(t, err)

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(deleteFile(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestApi_DeleteFile_ShouldReturn401IfErrorOccursValidatingToken(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	extHandler.On("ValidateToken", mock.Anything).Return(errors.New("test"))

	req, err := http.NewRequest(http.MethodDelete, "/file/5df25cc42d811e3b6b945c08", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(deleteFile(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestApi_DeleteFile_ShouldReturn400IfUnableToCreateObjectIDFromGivenIDVar(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	extHandler.On("ValidateToken", mock.Anything).Return(nil)

	req, err := http.NewRequest(http.MethodDelete, "/file/5df25cc42d811e3b6b945c08", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(deleteFile(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestApi_DeleteFile_ShouldReturn500OnDbHandlerError(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	dbHandler.On("DeleteFile", mock.Anything, mock.Anything).Return(errors.New("test"))
	extHandler.On("ValidateToken", mock.Anything).Return(nil)

	req, err := http.NewRequest(http.MethodDelete, "/file/5df25cc42d811e3b6b945c08", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req = mux.SetURLVars(req, map[string]string{"id": "5df25cc42d811e3b6b945c08"})

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(deleteFile(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestApi_DeleteFile_ShouldReturn200OnSuccess(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	dbHandler.On("DeleteFile", mock.Anything, mock.Anything).Return(nil)
	extHandler.On("ValidateToken", mock.Anything).Return(nil)

	req, err := http.NewRequest(http.MethodDelete, "/file/5df25cc42d811e3b6b945c08", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req = mux.SetURLVars(req, map[string]string{"id": "5df25cc42d811e3b6b945c08"})

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(deleteFile(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestApi_UpdateFileInfo_ShouldReturn400OnNoAuthorizationTokenFound(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}

	req, err := http.NewRequest(http.MethodPut, "/file/5df25cc42d811e3b6b945c08", nil)
	require.Nil(t, err)

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(updateFileInfo(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestApi_UpdateFileInfo_ShouldReturn401IfErrorOccursValidatingToken(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	extHandler.On("ValidateToken", mock.Anything).Return(errors.New("test"))

	req, err := http.NewRequest(http.MethodPut, "/file/5df25cc42d811e3b6b945c08", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(updateFileInfo(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestApi_UpdateFileInfo_ShouldReturn400IfUnableToCreateObjectIDFromGivenIDVar(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	extHandler.On("ValidateToken", mock.Anything).Return(nil)

	req, err := http.NewRequest(http.MethodPut, "/file/5df25cc42d811e3b6b945c08", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(updateFileInfo(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestApi_UpdateFileInfo_ShouldReturn400IfErrorsOccursDecodingRequestBody(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	extHandler.On("ValidateToken", mock.Anything).Return(nil)

	req, err := http.NewRequest(http.MethodPut, "/file/5df25cc42d811e3b6b945c08", strings.NewReader(""))
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req = mux.SetURLVars(req, map[string]string{"id": "5df25cc42d811e3b6b945c08"})

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(updateFileInfo(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestApi_UpdateFileInfo_ShouldReturn500OnDbHandlerError(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	dbHandler.On("UpdateFileInfo", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("test"))
	extHandler.On("ValidateToken", mock.Anything).Return(nil)

	req, err := http.NewRequest(http.MethodPut, "/file/5df25cc42d811e3b6b945c08", strings.NewReader("{}"))
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req = mux.SetURLVars(req, map[string]string{"id": "5df25cc42d811e3b6b945c08"})

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(updateFileInfo(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestApi_UpdateFileInfo_ShouldReturn200OnSuccess(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	dbHandler.On("UpdateFileInfo", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	extHandler.On("ValidateToken", mock.Anything).Return(nil)

	req, err := http.NewRequest(http.MethodPut, "/file/5df25cc42d811e3b6b945c08", strings.NewReader("{}"))
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req = mux.SetURLVars(req, map[string]string{"id": "5df25cc42d811e3b6b945c08"})

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(updateFileInfo(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestApi_GetFiles_ShouldReturn400OnNoAuthorizationTokenFound(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}

	req, err := http.NewRequest(http.MethodGet, "/files", nil)
	require.Nil(t, err)

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(getFiles(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestApi_GetFiles_ShouldReturn401IfErrorOccursValidatingToken(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	extHandler.On("ValidateToken", mock.Anything).Return(errors.New("test"))

	req, err := http.NewRequest(http.MethodGet, "/files", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(getFiles(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestApi_GetFiles_ShouldReturn500OnDbHandlerError(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	dbHandler.On("GetFiles", mock.Anything, mock.Anything).Return(nil, errors.New("test"))
	extHandler.On("ValidateToken", mock.Anything).Return(nil)

	req, err := http.NewRequest(http.MethodGet, "/files?size=test", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(getFiles(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestApi_GetFiles_ShouldReturn200OnSuccess(t *testing.T) {
	dbHandler := &mocks.DBHandler{}
	extHandler := &mocks.ExtHandler{}
	dbHandler.On("GetFiles", mock.Anything, mock.Anything).Return([]models.FileResponse{{}}, nil)
	extHandler.On("ValidateToken", mock.Anything).Return(nil)

	req, err := http.NewRequest(http.MethodGet, "/files?size=1234&test=test", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer test")

	recorder := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(getFiles(dbHandler, extHandler))
	httpHandler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)
}
