package database

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const PERMISSION_COLLECTION = "permissions"

func GrantGlobalPermission(Email *string, Permission *int) bool {
	user, err := FindUserByEmail(Email)
	if err != nil {
		slog.Debug("Couldn't grant permission to user", "user", Email, "error", err)
		return false
	}

	filter := bson.M{"userId": user.Id.Hex()}
	opts := options.Update().SetUpsert(true)
	update := bson.M{"$addToSet": bson.M{"globalPermissions": Permission}}
	result, err := Mongo().Collection(PERMISSION_COLLECTION).UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		slog.Debug("Couldn't grant permission to user", "user", Email, "error", err, "permission", Permission)
		return false
	}
	slog.Debug("Permission granted", "email", Email, "permission", Permission, "result", result)
	return true
}

func RevokeGlobalPermission(Email *string, Permission *int) bool {
	user, err := FindUserByEmail(Email)
	if err != nil {
		slog.Debug("Couldn't revoke permission to user", "user", Email, "error", err)
		return false
	}

	filter := bson.M{"userId": user.Id.Hex()}
	update := bson.M{"$pull": bson.M{"globalPermissions": Permission}}
	result, err := Mongo().Collection(PERMISSION_COLLECTION).UpdateOne(context.Background(), filter, update, nil)
	if err != nil {
		slog.Debug("Couldn't revoke permission to user", "user", Email, "error", err, "permission", Permission)
		return false
	}
	slog.Debug("Permission revoked", "email", Email, "permission", Permission, "resutlt", result)
	return true
}
