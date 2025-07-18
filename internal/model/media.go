package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Media struct {
	Id primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	// The original filename (keep it when user download its files back)
	OriginalFileName *string `bson:"originalFileName" json:"originalFileName"`
	// Storage file name, the original file (full quality) storage file name
	StorageFileName *string `bson:"storageFileName" json:"storageFileName"`
	// Name of the compressed of this file (if any), stored in the compressed medias folder
	CompressedFileName *string `bson:"compressedFileName" json:"compressedFileName"`
	// The ID of the uploader
	UploadedBy *primitive.ObjectID `bson:"uploadedBy" json:"uploadedBy"`
	// The date time at which the file was uploaded
	UploadTime *time.Time `bson:"uploadTime" json:"uploadTime"`
	// Indicate if the file was uploaded via a shared link
	UploadedViaSharedLink bool `bson:"uploadedViaSharedLink" json:"uploadedViaSharedLink"`
	// Hash of the media data to ensure uniqueness
	Hash *string `bson:"hash" json:"hash"`
}
