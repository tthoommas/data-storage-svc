package repository

import (
	"context"
	"data-storage-svc/internal/model"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MediaRepository interface {
	// Create a new media resource in DB
	Create(media *model.Media) (*primitive.ObjectID, error)
	// Get a media by ID from DB
	Get(mediaId *primitive.ObjectID) (*model.Media, error)
	// Get all media uploaded by a given user
	GetAllUploadedBy(userId *primitive.ObjectID) ([]model.Media, error)
	// Delete a media from media collection only (will not delete underlying file or any other link!)
	Delete(mediaId *primitive.ObjectID) error
	// Update a media
	Update(mediaId *primitive.ObjectID, update bson.M) error
}

type mediaRepository struct {
	db *mongo.Database
}

const (
	MEDIA_COLLECTION = "medias"
)

func NewMediaRepository(db *mongo.Database) mediaRepository {
	return mediaRepository{db}
}

func (r mediaRepository) Create(media *model.Media) (*primitive.ObjectID, error) {
	result, err := r.db.Collection(MEDIA_COLLECTION).InsertOne(context.Background(), media)
	if err != nil {
		return nil, err
	}
	generatedId := result.InsertedID.(primitive.ObjectID)
	return &generatedId, nil
}

func (r mediaRepository) Get(mediaId *primitive.ObjectID) (*model.Media, error) {
	filter := bson.M{"_id": mediaId}
	singleResult := r.db.Collection(MEDIA_COLLECTION).FindOne(context.Background(), filter)
	var media model.Media
	err := singleResult.Decode(&media)
	if err != nil {
		return nil, err
	}
	return &media, err
}

func (r mediaRepository) GetAllUploadedBy(userId *primitive.ObjectID) ([]model.Media, error) {
	filter := bson.M{"uploadedBy": userId}
	cursor, err := r.db.Collection(MEDIA_COLLECTION).Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

	var medias []model.Media = make([]model.Media, 0)
	for cursor.Next(context.Background()) {
		var media model.Media
		if err = cursor.Decode(&media); err != nil {
			return nil, fmt.Errorf("unable to decode media from databas")
		} else {
			medias = append(medias, media)
		}
	}
	return medias, nil
}

func (r mediaRepository) Delete(mediaId *primitive.ObjectID) error {
	filter := bson.M{"_id": mediaId}
	_, err := r.db.Collection(MEDIA_COLLECTION).DeleteOne(context.Background(), filter)
	return err
}

func (r mediaRepository) Update(mediaId *primitive.ObjectID, update bson.M) error {
	filter := bson.M{"_id": mediaId}
	updateDoc := bson.M{"$set": update}

	_, err := r.db.Collection(MEDIA_COLLECTION).UpdateOne(context.Background(), filter, updateDoc)
	return err
}
