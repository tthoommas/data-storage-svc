package utils

import (
	"data-storage-svc/internal"
	"data-storage-svc/internal/model"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetDataDir(subPath string) (string, error) {
	path := filepath.Join(internal.DATA_DIR_PATH, subPath)
	exists, err := pathExists(path)
	if err != nil || !exists {
		slog.Debug("Data path not found", "path", path)
		return "", fmt.Errorf("couldn't open data folder, or path does not exists")
	}
	return path, nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil // path exists
	}
	if os.IsNotExist(err) {
		return false, nil // path does not exist
	}
	return false, err // some other error
}

// Decode a primitive.ObjectId from a query parameter
func DecodeQueryId(queryParam string, request *gin.Context) (*primitive.ObjectID, error) {
	rawId := request.Query(queryParam)
	if rawId == "" {
		slog.Debug("Unable to decode query ID. Empty query param.", "queryParam", queryParam)
		request.AbortWithStatus(http.StatusBadRequest)
		return nil, fmt.Errorf("query param [%s] is empty", queryParam)
	}
	decodedId, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		slog.Debug("Unable to decode query ID", "queryParam", queryParam, "error", err)
		request.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return nil, fmt.Errorf("unable to decode query id [%s], error = %s", queryParam, err)
	}
	return &decodedId, nil
}

// Decode a primitive.ObjectId from a request body
func DecodeBodyId(rawId string) (*primitive.ObjectID, ServiceError) {
	decodedId, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		slog.Debug("Unable to decode ID from body", "error", err)
		return nil, NewServiceError(http.StatusBadRequest, "unable to decode provided id")
	}
	return &decodedId, nil
}

func GetUser(request *gin.Context) (*model.User, error) {
	rawUser, _ := request.Get("user")
	if user, ok := rawUser.(*model.User); !ok {
		request.AbortWithStatus(http.StatusInternalServerError)
		return nil, errors.New("couldn't find user")
	} else {
		return user, nil
	}
}
