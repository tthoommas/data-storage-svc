package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Media struct {
	Id               primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	OriginalFileName *string             `bson:"originalFileName" json:"originalFileName"`
	StorageFileName  *string             `bson:"storageFileName" json:"storageFileName"`
	UploadedBy       *primitive.ObjectID `bson:"uploadedBy" json:"uploadedBy"`
	UploadTime       *time.Time          `bson:"uploadTime" json:"uploadTime"`
}
