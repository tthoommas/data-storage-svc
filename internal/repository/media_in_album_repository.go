package repository

import (
	"context"
	"data-storage-svc/internal/model"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Manage mediaInAblum resources that represents relations between medias and albums
type MediaInAlbumRepository interface {
	// Add an existing media in an existing album
	AddMediaToAlbum(mediaInAlbum *model.MediaInAlbum) error
	// Remove a media from an album
	RemoveMediaFromAlbum(albumId *primitive.ObjectID, mediaId *primitive.ObjectID) error
	// Remove a media from all albums
	RemoveMediaFromAllAlbums(mediaId *primitive.ObjectID) error
	// Unlink an album from all medias (i.e. remove any medias from the given album)
	UnlinkAlbumFromAllMedias(albumId *primitive.ObjectID) error
	// List all medias in an album
	ListAllMedias(albumId *primitive.ObjectID) ([]model.MediaInAlbum, error)
}

type mediaInAlbumRepository struct {
	db *mongo.Database
}

func NewMediaInAlbumRepository(db *mongo.Database) MediaInAlbumRepository {
	return mediaInAlbumRepository{db}
}

func (r mediaInAlbumRepository) AddMediaToAlbum(mediaInAlbum *model.MediaInAlbum) error {
	_, err := r.db.Collection(MEDIA_IN_ALBUM_COLLECTION).InsertOne(context.Background(), mediaInAlbum)
	return err
}

func (r mediaInAlbumRepository) RemoveMediaFromAlbum(albumId *primitive.ObjectID, mediaId *primitive.ObjectID) error {
	filter := bson.M{"mediaId": mediaId, "albumId": albumId}
	_, err := r.db.Collection(MEDIA_IN_ALBUM_COLLECTION).DeleteMany(context.Background(), filter)
	return err
}

func (r mediaInAlbumRepository) RemoveMediaFromAllAlbums(mediaId *primitive.ObjectID) error {
	filter := bson.M{"mediaId": mediaId}
	_, err := r.db.Collection(MEDIA_IN_ALBUM_COLLECTION).DeleteMany(context.Background(), filter)
	return err
}

func (r mediaInAlbumRepository) UnlinkAlbumFromAllMedias(albumId *primitive.ObjectID) error {
	filter := bson.M{"albumId": albumId}
	_, err := r.db.Collection(MEDIA_IN_ALBUM_COLLECTION).DeleteMany(context.Background(), filter)
	return err
}

func (r mediaInAlbumRepository) ListAllMedias(albumId *primitive.ObjectID) ([]model.MediaInAlbum, error) {
	filter := bson.M{"albumId": albumId}
	cursor, err := r.db.Collection(MEDIA_IN_ALBUM_COLLECTION).Find(context.Background(), filter)
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
