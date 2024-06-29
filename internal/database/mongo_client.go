package database

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DB_NAME = "db"
)

var mongoClient *mongo.Client

func Mongo() *mongo.Database {
	if mongoClient == nil {
		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
		if err != nil {
			slog.Error("could create mongo client")
			panic(err)
		}
		configureMongoDb(client)
		mongoClient = client
	}
	return mongoClient.Database(DB_NAME)
}

func configureMongoDb(client *mongo.Client) {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true), // Unique index
	}
	client.Database(DB_NAME).Collection(USER_COLLECTION).Indexes().CreateOne(context.Background(), indexModel)
}
