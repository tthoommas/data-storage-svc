package repository

import (
	"context"
	"data-storage-svc/internal/model"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MediaAccessRepository interface {
	// Create a media access for a user and a media
	Create(userId *primitive.ObjectID, mediaId *primitive.ObjectID) error
	// Delete a media access for a user and a media
	Remove(userId *primitive.ObjectID, mediaId *primitive.ObjectID) error
	// Delete all media accesses for a given media
	RemoveAll(mediaId *primitive.ObjectID) error
	// Get all media accesses associated to a given user
	GetAllForUser(userId *primitive.ObjectID) ([]model.UserMediaAccess, error)
	// Get a media access (if it exists) from user and media
	Get(userId *primitive.ObjectID, mediaId *primitive.ObjectID) (*model.UserMediaAccess, error)
}

type mediaAccessRepository struct {
	db *mongo.Database
}

const USER_MEDIA_ACCESS_COLLECTION = "users_media_access"

func NewMediaAccessRepository(db *mongo.Database) mediaAccessRepository {
	return mediaAccessRepository{db}
}

func (r mediaAccessRepository) Create(UserId *primitive.ObjectID, MediaId *primitive.ObjectID) error {
	access := model.UserMediaAccess{UserId: UserId, MediaId: MediaId}
	_, err := r.db.Collection(USER_MEDIA_ACCESS_COLLECTION).InsertOne(context.Background(), access)
	return err
}

func (r mediaAccessRepository) Remove(UserId *primitive.ObjectID, MediaId *primitive.ObjectID) error {
	filter := bson.M{"userId": *UserId, "mediaId": *MediaId}
	_, err := r.db.Collection(USER_MEDIA_ACCESS_COLLECTION).DeleteOne(context.Background(), filter)
	return err
}

func (r mediaAccessRepository) RemoveAll(mediaId *primitive.ObjectID) error {
	filter := bson.M{"mediaId": mediaId}
	_, err := r.db.Collection(USER_MEDIA_ACCESS_COLLECTION).DeleteMany(context.Background(), filter)
	return err
}

func (r mediaAccessRepository) GetAllForUser(userId *primitive.ObjectID) ([]model.UserMediaAccess, error) {
	filter := bson.D{{Key: "userId", Value: userId}}
	cursor, err := r.db.Collection(USER_MEDIA_ACCESS_COLLECTION).Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var mediaAccesses []model.UserMediaAccess = make([]model.UserMediaAccess, 0)
	for cursor.Next(context.Background()) {
		var mediaAccess model.UserMediaAccess
		if err = cursor.Decode(&mediaAccess); err != nil {
			slog.Error("Couldn't decode media id", "error", err)
		} else {
			mediaAccesses = append(mediaAccesses, mediaAccess)
		}
	}
	return mediaAccesses, nil
}

func (r mediaAccessRepository) Get(userId *primitive.ObjectID, mediaId *primitive.ObjectID) (*model.UserMediaAccess, error) {
	filter := bson.D{{Key: "userId", Value: userId}, {Key: "mediaId", Value: mediaId}}
	result := r.db.Collection(USER_MEDIA_ACCESS_COLLECTION).FindOne(context.Background(), filter)
	if result.Err() != nil {
		slog.Debug("couldn't find media. Cannot check if user is allowed to access it.", "error", result.Err(), "userId", userId.Hex(), "mediaId", mediaId.Hex())
		return nil, fmt.Errorf("couldn't find usermediaccess in database - error [%s]", result.Err())
	}
	var userMediaAccess *model.UserMediaAccess
	err := result.Decode(userMediaAccess)
	if err != nil {
		return nil, fmt.Errorf("couldn't decode usermediaaccess from database error - [%s]", err)
	}
	return userMediaAccess, nil
}
