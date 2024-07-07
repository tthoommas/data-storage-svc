package model

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	PERMISSION_LOGIN        = 0
	PERMISSION_UPLOAD_MEDIA = 1
)

type UserPermissions struct {
	UserId            primitive.ObjectID   `bson:"userId"`
	GlobalPermissions []int                `bson:"globalPermissions"`
	AllowedToView     []primitive.ObjectID `bson:"allowedToView"`
	AllowedToEdit     []primitive.ObjectID `bson:"allowedToEdit"`
}
