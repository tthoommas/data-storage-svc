package database

import (
	"context"
	"data-storage-svc/internal/model"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const USER_MEDIA_ACCESS_COLLECTION = "users_media_access"
const USER_ALBUM_ACCESS_COLLECTION = "users_album_access"

func GrantAccessToMedia(UserId *primitive.ObjectID, MediaId *primitive.ObjectID) error {
	access := model.UserMediaAccess{UserId: UserId, MediaId: MediaId}
	_, err := Mongo().Collection(USER_MEDIA_ACCESS_COLLECTION).InsertOne(context.Background(), access)
	return err
}

func GrantAccessToAlbum(UserId *primitive.ObjectID, AlbumId *primitive.ObjectID, CanEdit bool) error {
	filter := bson.M{
		"userId":  UserId,
		"albumId": AlbumId,
	}
	update := bson.M{
		"$set": bson.M{
			"canEdit": CanEdit,
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := Mongo().Collection(USER_ALBUM_ACCESS_COLLECTION).UpdateOne(context.Background(), filter, update, opts)
	return err
}

func RevokeAccessToAlbum(UserId *primitive.ObjectID, AlbumId *primitive.ObjectID) error {
	filter := bson.M{
		"userId":  UserId,
		"albumId": AlbumId,
	}
	_, err := Mongo().Collection(USER_ALBUM_ACCESS_COLLECTION).DeleteOne(context.Background(), filter)
	return err
}

func GetAllMediasForUser(UserId *primitive.ObjectID) ([]primitive.ObjectID, error) {
	filter := bson.D{{Key: "userId", Value: UserId}}
	cursor, err := Mongo().Collection(USER_MEDIA_ACCESS_COLLECTION).Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var mediaAccesses []primitive.ObjectID = make([]primitive.ObjectID, 0)
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

func CanUserAccessAlbum(UserId *primitive.ObjectID, AlbumId *primitive.ObjectID) bool {
	filter := bson.D{{Key: "userId", Value: UserId}, {Key: "albumId", Value: AlbumId}}
	result := Mongo().Collection(USER_ALBUM_ACCESS_COLLECTION).FindOne(context.Background(), filter)
	if result.Err() != nil {
		slog.Debug("couldn't find album access. User probably don't have permission to view this album", "error", result.Err(), "userId", UserId.Hex(), "albumId", AlbumId.Hex())
		return false
	}
	return true
}

func CanUserEditAlbum(UserId *primitive.ObjectID, AlbumId *primitive.ObjectID) bool {
	filter := bson.D{{Key: "userId", Value: UserId}, {Key: "albumId", Value: AlbumId}}
	result := Mongo().Collection(USER_ALBUM_ACCESS_COLLECTION).FindOne(context.Background(), filter)
	if result.Err() != nil {
		slog.Debug("couldn't find album access. User probably don't have permission to view (and thus edit) this album", "error", result.Err(), "userId", UserId.Hex(), "albumId", AlbumId.Hex())
		return false
	}
	var access model.UserAlbumAccess
	err := result.Decode(&access)
	if err != nil {
		slog.Debug("Couldn't decode user album access from DB", "error", err)
		return false
	}
	return access.CanEdit
}

func GetAllAlbumsForUser(UserId *primitive.ObjectID) ([]model.UserAlbumAccess, error) {
	filter := bson.M{"userId": UserId}
	cursor, err := Mongo().Collection(USER_ALBUM_ACCESS_COLLECTION).Find(context.Background(), filter)
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
