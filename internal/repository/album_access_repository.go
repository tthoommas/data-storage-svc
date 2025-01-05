package repository

import (
	"context"
	"data-storage-svc/internal/model"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AlbumAccessRepository interface {
	// Create an album access entry to a given album to for a given user
	Create(userId *primitive.ObjectID, albumId *primitive.ObjectID, canEdit bool) error
	// Remove an album access entry to a given album for a given user
	Remove(userId *primitive.ObjectID, albumId *primitive.ObjectID) error
	// Remove all album access entries for all users (after that, no body will be able to access/edit the album)
	RemoveAllAccesses(albumId *primitive.ObjectID) error
	// Get all album accesses associated to a given user id
	GetAllByUser(userId *primitive.ObjectID) ([]model.UserAlbumAccess, error)
	// Get all album accesses associated to a given album id
	GetAllByAlbum(albumId *primitive.ObjectID) ([]model.UserAlbumAccess, error)
	// Get a specific album access for a given userId and albumId
	Get(userId *primitive.ObjectID, albumId *primitive.ObjectID) (*model.UserAlbumAccess, error)
}

const USER_ALBUM_ACCESS_COLLECTION = "users_album_access"

type albumAccessRepository struct {
	db *mongo.Database
}

func NewAlbumAccessRepository(db *mongo.Database) albumAccessRepository {
	return albumAccessRepository{db}
}

func (r albumAccessRepository) Create(userId *primitive.ObjectID, albumId *primitive.ObjectID, canEdit bool) error {
	filter := bson.M{
		"userId":  userId,
		"albumId": albumId,
	}
	update := bson.M{
		"$set": bson.M{
			"canEdit": canEdit,
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := r.db.Collection(USER_ALBUM_ACCESS_COLLECTION).UpdateOne(context.Background(), filter, update, opts)
	return err
}

func (r albumAccessRepository) Remove(userId *primitive.ObjectID, albumId *primitive.ObjectID) error {
	filter := bson.M{
		"userId":  userId,
		"albumId": albumId,
	}
	_, err := r.db.Collection(USER_ALBUM_ACCESS_COLLECTION).DeleteOne(context.Background(), filter)
	return err
}

func (r albumAccessRepository) RemoveAllAccesses(albumId *primitive.ObjectID) error {
	filter := bson.M{
		"albumId": albumId,
	}
	_, err := r.db.Collection(USER_ALBUM_ACCESS_COLLECTION).DeleteMany(context.Background(), filter)
	return err
}

func (r albumAccessRepository) GetAllByUser(userId *primitive.ObjectID) ([]model.UserAlbumAccess, error) {
	filter := bson.M{"userId": userId}
	cursor, err := r.db.Collection(USER_ALBUM_ACCESS_COLLECTION).Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

	var userAlbumAccesses []model.UserAlbumAccess = make([]model.UserAlbumAccess, 0)
	for cursor.Next(context.Background()) {
		var userAlbumAccess model.UserAlbumAccess
		if err = cursor.Decode(&userAlbumAccess); err != nil {
			slog.Error("Couldn't decode album", "error", err)
		} else {
			userAlbumAccesses = append(userAlbumAccesses, userAlbumAccess)
		}
	}

	return userAlbumAccesses, nil
}

func (r albumAccessRepository) Get(userId *primitive.ObjectID, albumId *primitive.ObjectID) (*model.UserAlbumAccess, error) {
	filter := bson.D{{Key: "userId", Value: userId}, {Key: "albumId", Value: albumId}}
	result := r.db.Collection(USER_ALBUM_ACCESS_COLLECTION).FindOne(context.Background(), filter)
	if result.Err() != nil {
		slog.Debug("couldn't find album access. User probably don't have permission to view this album", "error", result.Err(), "userId", userId.Hex(), "albumId", albumId.Hex())
		return nil, fmt.Errorf("couldn't find album access. User probably don't have permission to view this album - error [%s]", result.Err())
	}
	var albumAccess model.UserAlbumAccess
	err := result.Decode(&albumAccess)
	if err != nil {
		return nil, err
	}
	return &albumAccess, nil
}

func (r albumAccessRepository) GetAllByAlbum(albumId *primitive.ObjectID) ([]model.UserAlbumAccess, error) {
	filter := bson.M{"albumId": albumId}
	cursor, err := r.db.Collection(USER_ALBUM_ACCESS_COLLECTION).Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

	var userAlbumAccesses []model.UserAlbumAccess = make([]model.UserAlbumAccess, 0)
	for cursor.Next(context.Background()) {
		var userAlbumAccess model.UserAlbumAccess
		if err = cursor.Decode(&userAlbumAccess); err != nil {
			slog.Error("Couldn't decode album", "error", err)
		} else {
			userAlbumAccesses = append(userAlbumAccesses, userAlbumAccess)
		}
	}

	return userAlbumAccesses, nil
}
