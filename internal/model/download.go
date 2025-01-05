package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Download struct {
	Id                      *primitive.ObjectID `bson:"_id,omitempty"`
	DownloadName            string              `bson:"downloadName" json:"downloadName"`
	StartedAt               *time.Time          `bson:"startedAt" json:"startedAt"`
	ZipFileName             *string             `bson:"zipFileName" json:"zipFileName"`
	IsReady                 bool                `bson:"isReady" json:"isReady"`
	Initiator               *primitive.ObjectID `bson:"initiator" json:"initiator"`
	IsInitiatedBySharedLink bool                `bson:"isInitBySharedLink" json:"isInitBySharedLink"`
}
