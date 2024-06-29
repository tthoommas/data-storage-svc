package endpoints

import (
	"data-storage-svc/internal/api/utils"
	"data-storage-svc/internal/database"
	"data-storage-svc/internal/model"
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
		slog.Debug("Not file name provided")
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

	uploadedBy := c.GetString("email")
	uploadTime := time.Now()
	err = database.CreateMedia(&model.Media{OriginalFileName: &fileName, StorageFileName: &storageFileName, UploadedBy: &uploadedBy, UploadTime: &uploadTime})
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

	c.Status(http.StatusOK)
}

var AllowedMediaExtension = map[string]bool{
	"jpg":  true,
	"jpeg": true,
	"png":  true,
	"gif":  true,
}

func checkExtension(filename *string) (string, bool) {
	parts := strings.Split(*filename, ".")
	extension := ""
	if len(parts) < 2 {
		slog.Debug("Invalid filename, no extension")
		return "", false
	} else {
		extension = parts[len(parts)-1]
	}

	isValid := AllowedMediaExtension[extension]
	slog.Debug("Examining file extension", "extension", extension, "accepted", isValid)
	return extension, isValid
}
