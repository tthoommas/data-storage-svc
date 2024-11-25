package database

import (
	"context"
	"data-storage-svc/internal/model"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ALBUM_COLLECTION          = "albums"
	MEDIA_IN_ALBUM_COLLECTION = "media_in_album"
)

func GetAlbumById(AlbumId *primitive.ObjectID) (*model.Album, error) {
	filter := bson.M{"_id": AlbumId}
	result := Mongo().Collection(ALBUM_COLLECTION).FindOne(context.Background(), filter)
	var album model.Album
	err := result.Decode(&album)
	if err != nil {
		return nil, err
	}
	return &album, err
}

func CreateAlbum(Album *model.Album) (*primitive.ObjectID, error) {
	result, err := Mongo().Collection(ALBUM_COLLECTION).InsertOne(context.Background(), Album)
	if err != nil {
		return nil, err
	}
	generatedId := result.InsertedID.(primitive.ObjectID)
	return &generatedId, nil
}

func AddMediaToAlbum(MediaInAlbum *model.MediaInAlbum) error {
	_, err := Mongo().Collection(MEDIA_IN_ALBUM_COLLECTION).InsertOne(context.Background(), MediaInAlbum)
	return err
}

func GetAllMediasInAlbum(AlbumId *primitive.ObjectID) ([]model.MediaInAlbum, error) {
	filter := bson.M{"albumId": AlbumId}
	cursor, err := Mongo().Collection(MEDIA_IN_ALBUM_COLLECTION).Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

	var mediasInAlbum []model.MediaInAlbum = make([]model.MediaInAlbum, 0)
	for cursor.Next(context.Background()) {
		var mediaInAlbum model.MediaInAlbum
		if err = cursor.Decode(&mediaInAlbum); err != nil {
			slog.Error("Couldn't decode album", "error", err)
		} else {
			mediasInAlbum = append(mediasInAlbum, mediaInAlbum)
		}
	}
	return mediasInAlbum, nil
}

func DeleteAlbum(AlbumId *primitive.ObjectID) error {

}
