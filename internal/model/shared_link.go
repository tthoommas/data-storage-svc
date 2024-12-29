package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SharedLink struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AlbumId        primitive.ObjectID `json:"albumId" bson:"albumId"`
	CreatedBy      primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	CreatedAt      time.Time          `json:"created" bson:"created"`
	Token          string             `json:"token" bson:"token"`
	ExpirationDate time.Time          `json:"expiration" bson:"expiration"`
	CanEdit        bool               `json:"canEdit" bson:"canEdit"`
}
