package utils

import (
	"bytes"
	"context"
	"data-storage-svc/internal"
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/model"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func GetDataDir(subPath string) (string, error) {
	path := filepath.Join(internal.DATA_DIRECTORY, subPath)
	exists, err := pathExists(path)
	if err != nil {
		return "", err
	}
	if !exists {
		if err := os.Mkdir(path, 0755); err != nil {
			return "", err
		}
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
		request.AbortWithStatus(http.StatusUnauthorized)
		return nil, errors.New("couldn't find user")
	} else {
		return user, nil
	}
}

func GetUserOrSharedLink(request *gin.Context) (*model.User, *model.SharedLink, error) {
	rawUser, userExists := request.Get(common.USER)
	rawSharedLink, linkExists := request.Get(common.SHARED_LINK)
	if !userExists && !linkExists {
		// No authenticated user and no link => unauthorized
		request.AbortWithStatus(http.StatusUnauthorized)
		return nil, nil, errors.New("couldn't find user nor token")
	}
	var user *model.User = nil
	var sharedLink *model.SharedLink = nil

	if userExists {
		if typedUser, ok := rawUser.(*model.User); ok {
			user = typedUser
		}
	}

	if linkExists {
		if typedLink, ok := rawSharedLink.(*model.SharedLink); ok {
			sharedLink = typedLink
		}
	}
	return user, sharedLink, nil
}

func GetUserOrSharedLinkGeneric(context context.Context) (*model.User, *model.SharedLink, error) {
	rawUser := context.Value(common.USER)
	rawSharedLink := context.Value(common.SHARED_LINK)
	if rawUser == nil && rawSharedLink == nil {
		return nil, nil, errors.New("couldn't find user nor token")
	}
	var user *model.User = nil
	var sharedLink *model.SharedLink = nil

	if rawUser != nil {
		if typedUser, ok := rawUser.(*model.User); ok {
			user = typedUser
		}
	}

	if rawSharedLink != nil {
		if typedLink, ok := rawSharedLink.(*model.SharedLink); ok {
			sharedLink = typedLink
		}
	}
	return user, sharedLink, nil
}

func GetIdFromContext(key string, c *gin.Context) primitive.ObjectID {
	id, _ := c.Get(key)
	return id.(primitive.ObjectID)
}

func GetUserIdOrLinkId(user *model.User, sharedLink *model.SharedLink) (*primitive.ObjectID, bool, error) {
	if user != nil {
		return &user.Id, false, nil
	} else if sharedLink != nil {
		return &sharedLink.Id, true, nil
	} else {
		return nil, false, fmt.Errorf("unable to get id from user of shared link, both are nil")
	}
}

func ToUTF8(input string) (string, error) {
	reader := bytes.NewReader([]byte(input))
	transformer := charmap.ISO8859_1.NewDecoder()
	utf8Reader := transform.NewReader(reader, transformer)
	utf8Bytes, err := io.ReadAll(utf8Reader)
	if err != nil {
		return "", err
	}
	return string(utf8Bytes), nil
}

func IntPtr(i int) *int {
	return &i
}

func StrPtr(str string) *string {
	return &str
}

func Ptr[T any](v T) *T {
	return &v
}
