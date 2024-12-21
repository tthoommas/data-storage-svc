package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Download struct {
	Id           *primitive.ObjectID `bson:"_id,omitempty"`
	DownloadName string              `bson:"downloadName"`
	StartedAt    *time.Time          `bson:"startedAt"`
	ZipFileName  *string             `bson:"zipFileName"`
	IsReady      bool                `bson:"isReady"`
	Initiator    *primitive.ObjectID `bson:"initator"`
}
