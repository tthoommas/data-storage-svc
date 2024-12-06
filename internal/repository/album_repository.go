package repository

import (
	"context"
	"data-storage-svc/internal/model"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AlbumRepository interface {
	// Retrieve a specific album from its ID
	GetById(id primitive.ObjectID) (*model.Album, error)
	// Create a new album resource in the DB
	Create(album *model.Album) (*primitive.ObjectID, error)
	// Update an existing album in the DB
	Update(album *model.Album) error
	// Delete an existing album in the DB
	Delete(albumId *primitive.ObjectID) error
}

const (
	ALBUM_COLLECTION          = "albums"
	MEDIA_IN_ALBUM_COLLECTION = "media_in_album"
)

type albumRepository struct {
	db *mongo.Database
}

func NewAlbumRepository(db *mongo.Database) albumRepository {
	return albumRepository{db}
}

func (r albumRepository) GetById(id primitive.ObjectID) (*model.Album, error) {
	filter := bson.M{"_id": id}
	result := r.db.Collection(ALBUM_COLLECTION).FindOne(context.Background(), filter)
	var album model.Album
	err := result.Decode(&album)
	if err != nil {
		return nil, err
	}
	return &album, err
}

func (r albumRepository) Create(Album *model.Album) (*primitive.ObjectID, error) {
	result, err := r.db.Collection(ALBUM_COLLECTION).InsertOne(context.Background(), Album)
	if err != nil {
		return nil, err
	}
	generatedId := result.InsertedID.(primitive.ObjectID)
	return &generatedId, nil
}

func (r albumRepository) Update(album *model.Album) error {
	return errors.New("not yet implemented")
}

func (r albumRepository) Delete(albumId *primitive.ObjectID) error {
	filter := bson.M{"_id": albumId}
	_, err := r.db.Collection(ALBUM_COLLECTION).DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}
	return nil
}
