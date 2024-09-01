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
		Options: options.Index().SetUnique(true),
	}
	client.Database(DB_NAME).Collection(USER_COLLECTION).Indexes().CreateOne(context.Background(), indexModel)

	indexPermission := mongo.IndexModel{
		Keys:    bson.D{{Key: "userId", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	client.Database(DB_NAME).Collection(PERMISSION_COLLECTION).Indexes().CreateOne(context.Background(), indexPermission)
	// Create user id index for fast lookup of medias accessible by an user
	indexUserId := mongo.IndexModel{
		Keys: bson.D{{Key: "userId", Value: 1}},
	}
	client.Database(DB_NAME).Collection(USER_MEDIA_ACCESS_COLLECTION).Indexes().CreateOne(context.Background(), indexUserId)
	// Create a (user & media) id index for fast checking if a user can access a given media + ensure uniqueness of entries
	indexUserMediaId := mongo.IndexModel{
		Keys:    bson.D{{Key: "userId", Value: 1}, {Key: "mediaId", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	client.Database(DB_NAME).Collection(USER_MEDIA_ACCESS_COLLECTION).Indexes().CreateOne(context.Background(), indexUserMediaId)

	// Create an index to quicly get all albums accessible to a user
	indexUserAlbumId := mongo.IndexModel{
		Keys:    bson.D{{Key: "userId", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	client.Database(DB_NAME).Collection(USER_ALBUM_ACCESS_COLLECTION).Indexes().CreateOne(context.Background(), indexUserAlbumId)

	// Create an index to check quickly if a given user can access a given album + ensure uniqueness
	userAlbumAccessKey := mongo.IndexModel{
		Keys:    bson.D{{Key: "albumId", Value: 1}, {Key: "userId", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	client.Database(DB_NAME).Collection(USER_ALBUM_ACCESS_COLLECTION).Indexes().CreateOne(context.Background(), userAlbumAccessKey)

	// Create an index to quickly get all media from a given album
	indexAlbumId := mongo.IndexModel{
		Keys: bson.D{{Key: "albumId", Value: 1}},
	}
	client.Database(DB_NAME).Collection(MEDIA_IN_ALBUM_COLLECTION).Indexes().CreateOne(context.Background(), indexAlbumId)

	// Ensure a media can only be added once to a given album
	uniqueMediaInAlbum := mongo.IndexModel{
		Keys:    bson.D{{Key: "albumId", Value: 1}, {Key: "mediaId", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	client.Database(DB_NAME).Collection(MEDIA_IN_ALBUM_COLLECTION).Indexes().CreateOne(context.Background(), uniqueMediaInAlbum)

}
