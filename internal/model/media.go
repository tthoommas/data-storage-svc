package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Media struct {
	Id               primitive.ObjectID `bson:"_id,omitempty"`
	OriginalFileName *string            `bson:"originalFileName"`
	StorageFileName  *string            `bson:"storageFileName"`
	UploadedBy       *string            `bson:"uploadedBy"`
	UploadTime       *time.Time         `bson:"uploadTime"`
}
