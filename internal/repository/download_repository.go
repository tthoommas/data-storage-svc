package repository

import (
	"context"
	"data-storage-svc/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DOWNLOAD_COLLECTION = "downloads"
)

type DownloadRepository interface {
	// Create a new download resource in the DB
	Create(download *model.Download) (*primitive.ObjectID, error)
	// Update a download in the DB
	MarkAsReady(downloadId *primitive.ObjectID) error
	// Get a download by ID
	Get(downloadId *primitive.ObjectID) (*model.Download, error)
}

type downloadRepository struct {
	db *mongo.Database
}

func NewDownloadRepository(db *mongo.Database) downloadRepository {
	return downloadRepository{db}
}

func (r downloadRepository) Create(download *model.Download) (*primitive.ObjectID, error) {
	result, err := r.db.Collection(DOWNLOAD_COLLECTION).InsertOne(context.Background(), download)
	if err != nil {
		return nil, err
	}
	generatedId := result.InsertedID.(primitive.ObjectID)
	return &generatedId, nil
}

func (r downloadRepository) MarkAsReady(downloadId *primitive.ObjectID) error {
	filter := bson.M{"_id": downloadId}
	update := bson.M{"$set": bson.M{"isReady": true}}
	_, err := r.db.Collection(DOWNLOAD_COLLECTION).UpdateOne(context.Background(), filter, update)
	return err
}

func (r downloadRepository) Get(downloadId *primitive.ObjectID) (*model.Download, error) {
	filter := bson.M{"_id": downloadId}
	result := r.db.Collection(DOWNLOAD_COLLECTION).FindOne(context.Background(), filter)
	var download model.Download
	err := result.Decode(&download)
	if err != nil {
		return nil, err
	}
	return &download, err
}
