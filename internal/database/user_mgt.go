package database

import (
	"context"
	"data-storage-svc/internal/model"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	USER_COLLECTION = "users"
)

func FindUserByEmail(Email *string) (*model.User, error) {
	result := Mongo().Collection(USER_COLLECTION).FindOne(context.Background(), bson.M{"email": *Email})
	var user model.User
	err := result.Decode(&user)
	if err != nil {
		slog.Debug("Couldn't decode user from mongo DB ", "email", *Email, "err", err)
		return nil, err
	}
	return &user, nil
}

func StoreNewUser(Email *string, PasswordHash *string) error {
	_, err := Mongo().Collection(USER_COLLECTION).InsertOne(context.Background(), bson.M{"email": Email, "passwordHash": PasswordHash})
	return err
}

func UserExists(Email *string) bool {
	user, err := FindUserByEmail(Email)
	return err == nil && user != nil
}
