package database

import (
	"context"
	"data-storage-svc/internal/model"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const USER_MEDIA_ACCESS_COLLECTION = "users_media_access"

func GrantAccessToMedia(UserId *primitive.ObjectID, MediaId *primitive.ObjectID) error {
	access := model.UserMediaAccess{UserId: UserId, MediaId: MediaId}
	_, err := Mongo().Collection(USER_MEDIA_ACCESS_COLLECTION).InsertOne(context.Background(), access)
	return err
}

func GetAllMediasForUser(UserId *primitive.ObjectID) ([]primitive.ObjectID, error) {
	filter := bson.D{{Key: "userId", Value: UserId}}
	cursor, err := Mongo().Collection(USER_MEDIA_ACCESS_COLLECTION).Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var mediaAccesses []primitive.ObjectID
	for cursor.Next(context.Background()) {
		var mediaAccess model.UserMediaAccess
		if err = cursor.Decode(&mediaAccess); err != nil {
			slog.Error("Couldn't decode media id", "error", err)
		} else {
			mediaAccesses = append(mediaAccesses, *mediaAccess.MediaId)
		}
	}
	return mediaAccesses, nil
}

func CanUserAccessMedia(UserId *primitive.ObjectID, MediaId *primitive.ObjectID) bool {
	filter := bson.D{{Key: "userId", Value: UserId}, {Key: "mediaId", Value: MediaId}}
	result := Mongo().Collection(USER_MEDIA_ACCESS_COLLECTION).FindOne(context.Background(), filter)
	if result.Err() != nil {
		slog.Debug("couldn't find media. Cannot check if user is allowed to access it.", "error", result.Err(), "userId", UserId.Hex(), "mediaId", MediaId.Hex())
		return false
	}
	return true
}
