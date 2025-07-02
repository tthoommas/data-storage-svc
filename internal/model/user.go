package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	IsAdmin      bool               `bson:"isAdmin,omitempty" json:"isAdmin"`
	Email        string             `bson:"email" json:"email"`
	JoinDate     time.Time          `bson:"joinDate,omitempty" json:"joinDate"`
	LastLogin    time.Time          `bson:"lastLogin" json:"lastLogin"`
	PasswordHash string             `bson:"passwordHash,omitempty" json:"passwordHash,omitempty"`
}
