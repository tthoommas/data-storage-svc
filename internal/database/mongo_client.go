package database

import (
	"context"
	"data-storage-svc/internal/cli"
	"data-storage-svc/internal/repository"
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
			slog.Error("couldn't create mongo client")
			panic(err)
		}
		configureMongoDb(client, cli.DbName)
		mongoClient = client
	}

	return mongoClient.Database(cli.DbName)
}

func configureMongoDb(client *mongo.Client, dbName string) {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	client.Database(dbName).Collection(repository.USER_COLLECTION).Indexes().CreateOne(context.Background(), indexModel)
	// Create user id index for fast lookup of medias accessible by an user
	indexUserId := mongo.IndexModel{
		Keys: bson.D{{Key: "userId", Value: 1}},
	}
	client.Database(dbName).Collection(repository.USER_MEDIA_ACCESS_COLLECTION).Indexes().CreateOne(context.Background(), indexUserId)
	// Create a (user & media) id index for fast checking if a user can access a given media + ensure uniqueness of entries
	indexUserMediaId := mongo.IndexModel{
		Keys:    bson.D{{Key: "userId", Value: 1}, {Key: "mediaId", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	client.Database(dbName).Collection(repository.USER_MEDIA_ACCESS_COLLECTION).Indexes().CreateOne(context.Background(), indexUserMediaId)

	// Create an index to quicly get all albums accessible to a user
	indexUserAlbumId := mongo.IndexModel{
		Keys: bson.D{{Key: "userId", Value: 1}},
	}
	client.Database(dbName).Collection(repository.USER_ALBUM_ACCESS_COLLECTION).Indexes().CreateOne(context.Background(), indexUserAlbumId)

	// Create an index to check quickly if a given user can access a given album + ensure uniqueness
	userAlbumAccessKey := mongo.IndexModel{
		Keys:    bson.D{{Key: "albumId", Value: 1}, {Key: "userId", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	client.Database(dbName).Collection(repository.USER_ALBUM_ACCESS_COLLECTION).Indexes().CreateOne(context.Background(), userAlbumAccessKey)

	// Create an index to quickly get all media from a given album
	indexAlbumId := mongo.IndexModel{
		Keys: bson.D{{Key: "albumId", Value: 1}},
	}
	client.Database(dbName).Collection(repository.MEDIA_IN_ALBUM_COLLECTION).Indexes().CreateOne(context.Background(), indexAlbumId)

	// Ensure a media can only be added once to a given album
	uniqueMediaInAlbum := mongo.IndexModel{
		Keys:    bson.D{{Key: "albumId", Value: 1}, {Key: "mediaId", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	client.Database(dbName).Collection(repository.MEDIA_IN_ALBUM_COLLECTION).Indexes().CreateOne(context.Background(), uniqueMediaInAlbum)
}
