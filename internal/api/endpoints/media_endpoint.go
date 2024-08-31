package endpoints

import (
	"data-storage-svc/internal/api/utils"
	"data-storage-svc/internal/database"
	"data-storage-svc/internal/model"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func UploadMedia(c *gin.Context) {
	contentType := c.GetHeader("Content-Type")

	if contentType != "application/octet-stream" {
		slog.Debug("Invalid content type received while upload media", "content-type", contentType)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Expected Content-Type: application/octet-stream"})
		return
	}
	fileName := c.GetHeader("Content-Disposition")
	if len(fileName) == 0 {
		slog.Debug("No file name provided")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no file name provided while uploading"})
		return
	}
	extension, isValid := checkExtension(&fileName)
	if !isValid {
		slog.Debug("Reject file upload, invalid extension")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}
	targetFile, err := utils.GetDataDir("medias")
	if err != nil {
		slog.Debug("Couldn't upload media file", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "couldn't upload file"})
		return
	}
	storageFileName := uuid.NewString()
	storageFileName = storageFileName + "." + extension

	user, _ := c.Get("user")
	uploadedBy := user.(*model.User)
	uploadTime := time.Now()
	mediaId, err := database.CreateMedia(&model.Media{OriginalFileName: &fileName, StorageFileName: &storageFileName, UploadedBy: uploadedBy.Email, UploadTime: &uploadTime})
	if err != nil {
		slog.Debug("Impossible to add media into database", "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "couldn't upload file"})
		return
	}
	targetFile = filepath.Join(targetFile, storageFileName)
	outFile, err := os.Create(targetFile)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to create file: %s", err.Error()))
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, c.Request.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to save binary data: %s", err.Error()))
		return
	}

	err = database.GrantAccessToMedia(uploadedBy.Id, mediaId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "couldn't upload file"})
		slog.Debug("couldn't grant access to newly uploaded media", "error", err, "uploadedBy", *uploadedBy.Email, "mediaId", mediaId.Hex())
		return
	}

	c.Status(http.StatusOK)
}

var MediaExtensionMimeType = map[string]string{
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"gif":  "image/gif",
}

func getExtension(filename *string) (string, error) {
	parts := strings.Split(*filename, ".")
	extension := ""
	if len(parts) < 2 {
		slog.Debug("Invalid filename, no extension", "filename", filename)
		return "", errors.New("couldn't decode file extension")
	} else {
		extension = parts[len(parts)-1]
	}
	return extension, nil
}

func checkExtension(filename *string) (string, bool) {
	extension, err := getExtension(filename)
	if err != nil {
		return "", false
	}
	isValid := MediaExtensionMimeType[extension]
	slog.Debug("Examining file extension", "extension", extension, "accepted", isValid)
	return extension, isValid != ""
}

func getMimeType(filename *string) (string, error) {
	extension, err := getExtension(filename)
	if err != nil {
		return "", err
	}
	return MediaExtensionMimeType[extension], nil
}

func MyMedias(c *gin.Context) {
	rawUser, _ := c.Get("user")
	user := rawUser.(*model.User)
	medias, err := database.GetAllMediasForUser(user.Id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "couldn't fetch medias"})
		return
	}
	c.IndentedJSON(http.StatusOK, medias)
}

func GetMedia(c *gin.Context) {
	rawUser, _ := c.Get("user")
	user := rawUser.(*model.User)

	// Decode the media id requested in query param
	rawMediaId := c.Query("mediaId")
	mediaId, error := primitive.ObjectIDFromHex(rawMediaId)
	if rawMediaId == "" || error != nil {
		slog.Debug("Fetching media with invalid ID", "user", *user.Email, "rawMediaId", rawMediaId, "decodedMediaId", mediaId)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	// Check if user can access this media
	if !database.CanUserAccessMedia(user.Id, &mediaId) {
		slog.Debug("User is not allowed to access this media", "user", *user.Email, "media", mediaId.Hex())
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not allowed to access this resource"})
		return
	}

	// Get the media meta-info
	media, err := database.GetMedia(&mediaId)
	if err != nil {
		slog.Debug("Couldn't get media meta-infos", "error", err, "mediaId", mediaId)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "couldn't fetch media"})
		return
	}

	// Read media to send it
	mediaDirectory, err := utils.GetDataDir("medias")
	if err != nil {
		slog.Debug("Couldn't get data directory", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "couldn't fetch media"})
		return
	}
	mediaFilePath := filepath.Join(mediaDirectory, *media.StorageFileName)
	slog.Debug("Reading requested file", "filePath", mediaFilePath)
	data, err := os.ReadFile(mediaFilePath)
	if err != nil {
		slog.Debug("Cannot read requested media file", "error", err, "mediaFilePath", mediaFilePath)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "couldn't fetch media"})
		return
	}

	slog.Debug("read file", "length", len(data))

	// Infer mime type and send the media
	mimeType, err := getMimeType(media.StorageFileName)
	if err != nil {
		slog.Debug("Cannot infer file mime type", "error", err, "mediaFilePath", mediaFilePath)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "couldn't fetch media"})
		return
	}
	c.Data(http.StatusOK, mimeType, data)
}
