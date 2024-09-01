package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Album struct {
	Id           *primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title        string              `bson:"title" json:"title"`
	Description  string              `bson:"description" json:"description"`
	AuthorId     *primitive.ObjectID `bson:"authorId" json:"authorId"`
	CreationDate time.Time           `bson:"creationDate" json:"creationDate"`
}
