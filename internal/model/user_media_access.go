package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserMediaAccess struct {
	AccessId *primitive.ObjectID `bson:"_id,omitempty"`
	UserId   *primitive.ObjectID `bson:"userId"`
	MediaId  *primitive.ObjectID `bson:"mediaId"`
}
