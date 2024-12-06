package repository

import (
	"context"
	"data-storage-svc/internal/model"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
	// Create a new user resource in DB
	Create(user *model.User) (*primitive.ObjectID, error)
	// Get a specific user by email
	GetByEmail(email string) (*model.User, error)
	// Get a specific user by id
	GetById(id *primitive.ObjectID) (*model.User, error)
}

type userRepository struct {
	db *mongo.Database
}

func NewUserRepository(db *mongo.Database) userRepository {
	return userRepository{db}
}

const (
	USER_COLLECTION = "users"
)

func (r userRepository) Create(user *model.User) (*primitive.ObjectID, error) {
	_, err := r.db.Collection(USER_COLLECTION).InsertOne(context.Background(), user)
	return nil, err
}

func (r userRepository) GetByEmail(email string) (*model.User, error) {
	result := r.db.Collection(USER_COLLECTION).FindOne(context.Background(), bson.M{"email": email})
	var user model.User
	err := result.Decode(&user)
	if err != nil {
		slog.Debug("Couldn't decode user from mongo DB ", "email", email, "err", err)
		return nil, err
	}
	return &user, nil
}

func (r userRepository) GetById(id *primitive.ObjectID) (*model.User, error) {
	result := r.db.Collection(USER_COLLECTION).FindOne(context.Background(), bson.M{"_id": id})
	if result.Err() != nil {
		return nil, result.Err()
	}
	var user model.User
	err := result.Decode(&user)
	if err != nil {
		slog.Debug("Couldn't decode user from mongo DB ", "userId", *id, "err", err)
		return nil, err
	}
	return &user, nil
}
