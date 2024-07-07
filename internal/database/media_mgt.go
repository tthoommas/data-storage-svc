package database

import (
	"context"
	"data-storage-svc/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	MEDIA_COLLECTION = "medias"
)

func CreateMedia(Media *model.Media) (*primitive.ObjectID, error) {
	result, err := Mongo().Collection(MEDIA_COLLECTION).InsertOne(context.Background(), Media)
	if err != nil {
		return nil, err
	}
	generatedId := result.InsertedID.(primitive.ObjectID)
	return &generatedId, nil
}

func GetMedia(MediaId *primitive.ObjectID) (*model.Media, error) {
	filter := bson.M{"_id": MediaId}
	singleResult := Mongo().Collection(MEDIA_COLLECTION).FindOne(context.Background(), filter)
	var media model.Media
	err := singleResult.Decode(&media)
	if err != nil {
		return nil, err
	}
	return &media, err
}
