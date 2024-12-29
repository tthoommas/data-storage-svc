package repository

import (
	"context"
	"data-storage-svc/internal/model"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SharedLinkRepository interface {
	// Create a new shared link in DB
	Create(sharedLink *model.SharedLink) (*primitive.ObjectID, error)
	// Retrieve a shared link by ID
	Get(sharedLinkId *primitive.ObjectID) (*model.SharedLink, error)
	// Retrieve a shared link by its token
	GetByToken(sharedLinkToken string) (*model.SharedLink, error)
	// List all shared link for a given album id
	List(albumId *primitive.ObjectID) ([]model.SharedLink, error)
	// Delete a given shared link
	Delete(sharedLinkId primitive.ObjectID) error
	// Update a given link
	Update(sharedLinkId primitive.ObjectID, canEdit bool) error
}

type sharedLinkRepository struct {
	db *mongo.Database
}

const (
	SHARED_LINK_COLLECTION = "shared_links"
)

func NewSharedLinkRepository(db *mongo.Database) SharedLinkRepository {
	return sharedLinkRepository{db}
}

func (r sharedLinkRepository) Create(sharedLink *model.SharedLink) (*primitive.ObjectID, error) {
	result, err := r.db.Collection(SHARED_LINK_COLLECTION).InsertOne(context.Background(), sharedLink)
	if err != nil {
		return nil, err
	}
	generatedId := result.InsertedID.(primitive.ObjectID)
	return &generatedId, nil
}

func (r sharedLinkRepository) Get(sharedLinkId *primitive.ObjectID) (*model.SharedLink, error) {
	filter := bson.M{"_id": sharedLinkId}
	singleResult := r.db.Collection(SHARED_LINK_COLLECTION).FindOne(context.Background(), filter)
	var sharedLink model.SharedLink
	err := singleResult.Decode(&sharedLink)
	if err != nil {
		return nil, err
	}
	return &sharedLink, err
}

func (r sharedLinkRepository) GetByToken(sharedLinkToken string) (*model.SharedLink, error) {
	filter := bson.M{"token": sharedLinkToken}
	singleResult := r.db.Collection(SHARED_LINK_COLLECTION).FindOne(context.Background(), filter)
	var sharedLink model.SharedLink
	err := singleResult.Decode(&sharedLink)
	if err != nil {
		return nil, err
	}
	return &sharedLink, err
}

func (r sharedLinkRepository) List(albumId *primitive.ObjectID) ([]model.SharedLink, error) {
	filter := bson.M{"albumId": albumId}
	cursor, err := r.db.Collection(SHARED_LINK_COLLECTION).Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

	var sharedLinks []model.SharedLink = make([]model.SharedLink, 0)
	for cursor.Next(context.Background()) {
		var sharedLink model.SharedLink
		if err = cursor.Decode(&sharedLink); err != nil {
			slog.Error("Couldn't decode album", "error", err)
		} else {
			sharedLinks = append(sharedLinks, sharedLink)
		}
	}

	return sharedLinks, err
}

func (r sharedLinkRepository) Delete(sharedLinkId primitive.ObjectID) error {
	filter := bson.M{"_id": sharedLinkId}
	_, err := r.db.Collection(SHARED_LINK_COLLECTION).DeleteOne(context.Background(), filter)
	return err
}

func (r sharedLinkRepository) Update(sharedLinkId primitive.ObjectID, canEdit bool) error {
	update := bson.M{
		"$set": bson.M{
			"canEdit": canEdit,
		},
	}
	_, err := r.db.Collection(SHARED_LINK_COLLECTION).UpdateByID(context.Background(), sharedLinkId, update)
	return err
}
