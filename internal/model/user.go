package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id           *primitive.ObjectID `bson:"_id,omitempty"`
	IsAdmin      *bool               `bson:"isAdmin,omitempty"`
	Email        *string             `bson:"email"`
	JoinDate     *time.Time          `bson:"joinDate,omitempty"`
	PasswordHash *string             `bson:"passwordHash"`
}
