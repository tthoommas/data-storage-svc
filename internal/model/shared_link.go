package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SharedLink struct {
	Id        primitive.ObjectID `bson:"_id,omitempty"`
	AlbumId   primitive.ObjectID `bson:"albumId"`
	CreatedBy primitive.ObjectID `bson:"createdBy"`
	CreatedAt time.Time          `bson:"created"`
	Token     string             `bson:"token"`
}
