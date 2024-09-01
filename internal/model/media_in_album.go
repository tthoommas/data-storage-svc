package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MediaInAlbum struct {
	LinkId    *primitive.ObjectID `bson:"_id,omitempty"`
	MediaId   *primitive.ObjectID `bson:"mediaId"`
	AlbumId   *primitive.ObjectID `bson:"albumId"`
	AddedBy   *primitive.ObjectID `bson:"addedBy"`
	AddedDate *time.Time          `bson:"addedTime"`
}
