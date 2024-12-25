package repository

import (
	"context"
	"data-storage-svc/internal/model"

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
