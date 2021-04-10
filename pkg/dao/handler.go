package dao

import (
	"bytes"
	"context"
	"errors"

	"content-service-api/models"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DBHandler interface {
	Ping(ctx context.Context) error
	GetFile(ctx context.Context, fileID primitive.ObjectID) ([]byte, error)
	UploadFile(ctx context.Context, uploadRequest *models.FileRequest, fileBytes []byte) error
	DeleteFile(ctx context.Context, fileID primitive.ObjectID) error
	UpdateFileInfo(ctx context.Context, fileID primitive.ObjectID, updateRequest map[string]interface{}) error
	GetFiles(ctx context.Context, query map[string]interface{}) ([]models.FileResponse, error)
}

type Handler struct {
	Client          *mongo.Client
	Database        string
	FileCollection  string
	FsCollection    string
	ChunkCollection string
}

func (db *Handler) Ping(ctx context.Context) error {
	return db.Client.Ping(ctx, readpref.Primary())
}

func (db *Handler) GetFile(ctx context.Context, fileID primitive.ObjectID) ([]byte, error) {
	result := db.getFileCollection().FindOne(ctx, map[string]interface{}{"_id": fileID})
	if result.Err() != nil {
		return nil, result.Err()
	}

	var fileRequest models.FileRequest
	if err := result.Decode(&fileRequest); err != nil {
		return nil, err
	}

	bucket, err := gridfs.NewBucket(db.Client.Database(db.Database))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	_, err = bucket.DownloadToStream(fileRequest.FileID, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (db *Handler) UploadFile(ctx context.Context, uploadRequest *models.FileRequest, fileBytes []byte) error {
	bucket, err := gridfs.NewBucket(db.Client.Database(db.Database))
	if err != nil {
		return err
	}

	uploadStream, err := bucket.OpenUploadStream(uploadRequest.Name)
	if err != nil {
		return err
	}

	defer func() {
		if err := uploadStream.Close(); err != nil {
			logrus.WithError(err).Error("Error closing upload stream")
		}
	}()

	_, err = uploadStream.Write(fileBytes)
	if err != nil {
		return err
	}

	uploadRequest.FileID = uploadStream.FileID.(primitive.ObjectID)
	results, err := db.getFileCollection().InsertOne(ctx, uploadRequest)
	if err != nil {
		return err
	} else if results.InsertedID == nil {
		return errors.New("no file inserted")
	}

	return nil
}

func (db *Handler) DeleteFile(ctx context.Context, fileID primitive.ObjectID) error {
	result := db.getFileCollection().FindOneAndDelete(ctx, map[string]interface{}{"_id": fileID})
	if result.Err() != nil {
		return result.Err()
	}

	var fileRequest models.FileRequest
	if err := result.Decode(&fileRequest); err != nil {
		return err
	}

	logrus.Info(fileRequest)

	bucket, err := gridfs.NewBucket(db.Client.Database(db.Database))
	if err != nil {
		return err
	}

	if err = bucket.Delete(fileRequest.FileID); err != nil {
		return err
	}

	return nil
}

func (db *Handler) UpdateFileInfo(ctx context.Context, fileID primitive.ObjectID, updateRequest map[string]interface{}) error {
	updates := bson.M{}
	for key, val := range updateRequest {
		updates[key] = val
	}
	updates = bson.M{"$set": updates}

	result := db.getFileCollection().FindOneAndUpdate(ctx, map[string]interface{}{"_id": fileID}, updates)
	if result.Err() != nil {
		return result.Err()
	}

	return nil
}

func (db *Handler) GetFiles(ctx context.Context, query map[string]interface{}) ([]models.FileResponse, error) {
	cursor, err := db.getFileCollection().Find(ctx, query)
	if err != nil {
		return nil, err
	}

	var results []models.FileResponse
	err = cursor.All(ctx, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (db *Handler) getFileCollection() *mongo.Collection {
	return db.Client.Database(db.Database).Collection(db.FileCollection)
}

func (db *Handler) getFsCollection() *mongo.Collection {
	return db.Client.Database(db.Database).Collection(db.FsCollection)
}

func (db *Handler) getChunkCollection() *mongo.Collection {
	return db.Client.Database(db.Database).Collection(db.ChunkCollection)
}
