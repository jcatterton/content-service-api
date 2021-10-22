package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileRequest struct {
	Name      string             `json:"name" bson:"name"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
	Extension string             `json:"extension" bson:"extension"`
	Size      int64              `json:"size" bson:"size"`
	FileID    primitive.ObjectID `json:"fileBytes" bson:"fileBytes"`
	Hidden	  bool				 `json:"hidden" bson:"hidden"`
}

type FileUpdateRequest struct {
	Name      string    `json:"name" bson:"name"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	Extension string    `json:"extension" bson:"extension"`
	Size      int64     `json:"size" bson:"size"`
	Hidden	  bool	    `json:"hidden" bson:"hidden"`
}

type FileResponse struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Name      string             `json:"name" bson:"name"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
	Extension string             `json:"extension" bson:"extension"`
	Size      int64              `json:"size" bson:"size"`
	FileID    primitive.ObjectID `json:"fileBytes" bson:"fileBytes"`
	Hidden	  bool				 `json:"hidden" bson:"hidden"`
}
