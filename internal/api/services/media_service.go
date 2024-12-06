package services

import (
	"data-storage-svc/internal/model"
	"data-storage-svc/internal/repository"
	"data-storage-svc/internal/utils"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MediaService interface {
	// Create a new media resource (upload)
	Create(fileName string, uploader *primitive.ObjectID, data *io.ReadCloser) (*primitive.ObjectID, utils.ServiceError)
	// Get media by id
	GetById(mediaId *primitive.ObjectID) (*model.Media, utils.ServiceError)
	// Get the media data (i.e. bytes of the file stored on disk)
	GetData(storageFileName string) (*string, []byte, utils.ServiceError)
	// Get all media accessible to a given user
	GetAllSharedWithUser(userId *primitive.ObjectID) ([]model.UserMediaAccess, utils.ServiceError)
	// Get all medias uploaded by user
	GetAllUploadedByUser(userId *primitive.ObjectID) ([]model.Media, utils.ServiceError)
	// Delete a specific media
	Delete(mediaId *primitive.ObjectID) utils.ServiceError
}

type mediaService struct {
	// Repository dependencies
	mediaRepository repository.MediaRepository
	// Service dependencies
	mediaAccessService MediaAccessService
	albumService       AlbumService
}

func NewMediaService(mediaRepository repository.MediaRepository, mediaAccessService MediaAccessService, albumService AlbumService) mediaService {
	return mediaService{mediaRepository, mediaAccessService, albumService}
}

func (s mediaService) Create(fileName string, uploader *primitive.ObjectID, data *io.ReadCloser) (*primitive.ObjectID, utils.ServiceError) {
	if len(fileName) == 0 {
		return nil, utils.NewServiceError(http.StatusBadRequest, "no file name provided while uploading")
	}
	extension, isValid := checkExtension(&fileName)
	if !isValid {
		return nil, utils.NewServiceError(http.StatusBadRequest, "invalid filename")
	}
	targetFile, err := utils.GetDataDir("medias")
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't upload file")
	}
	storageFileName := uuid.NewString()
	storageFileName = storageFileName + "." + extension
	uploadTime := time.Now()
	mediaId, err := s.mediaRepository.Create(&model.Media{OriginalFileName: &fileName, StorageFileName: &storageFileName, UploadedBy: uploader, UploadTime: &uploadTime})
	if err != nil {
		return nil, utils.NewServiceError(http.StatusBadRequest, "couldn't upload file")
	}
	// Prepare to store file on disk
	targetFile = filepath.Join(targetFile, storageFileName)
	outFile, err := os.Create(targetFile)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't upload file")
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, *data)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "upload failed")
	}

	svcErr := s.mediaAccessService.GrantAccess(uploader, mediaId)
	if svcErr != nil {
		return nil, svcErr
	}
	return mediaId, nil
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

func (s mediaService) GetById(mediaId *primitive.ObjectID) (*model.Media, utils.ServiceError) {
	media, err := s.mediaRepository.Get(mediaId)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusNotFound, "media not found")
	}
	return media, nil
}

func (s mediaService) GetData(storageFileName string) (*string, []byte, utils.ServiceError) {
	// Read media to send it
	mediaDirectory, err := utils.GetDataDir("medias")
	if err != nil {
		return nil, nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't find media")
	}
	mediaFilePath := filepath.Join(mediaDirectory, storageFileName)
	slog.Debug("Reading requested file", "filePath", mediaFilePath)
	data, err := os.ReadFile(mediaFilePath)
	if err != nil {
		return nil, nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't find media")
	}

	slog.Debug("read file", "length", len(data))
	// Infer mime type and send the media
	mimeType, err := getMimeType(&storageFileName)
	if err != nil {
		slog.Debug("Cannot infer file mime type", "error", err, "mediaFilePath", mediaFilePath)
		return nil, nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't find media")
	}
	return &mimeType, data, nil
}

func (s mediaService) GetAllSharedWithUser(userId *primitive.ObjectID) ([]model.UserMediaAccess, utils.ServiceError) {
	return nil, utils.NewServiceError(http.StatusNotImplemented, "not yet implemented")
}

func (s mediaService) GetAllUploadedByUser(userId *primitive.ObjectID) ([]model.Media, utils.ServiceError) {
	medias, err := s.mediaRepository.GetAllUploadedBy(userId)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusNotFound, "no media found")
	}
	return medias, nil
}

func (s mediaService) Delete(mediaId *primitive.ObjectID) utils.ServiceError {
	mediaFolder, err := utils.GetDataDir("medias")
	if err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "couldn't delete media")
	}

	// Get the media meta-data
	media, svcErr := s.GetById(mediaId)
	if svcErr != nil {
		return svcErr
	}

	// Remove any user access to this media
	svcErr = s.mediaAccessService.RevokeAll(mediaId)
	if svcErr != nil {
		return svcErr
	}

	// Remove link to any album
	svcErr = s.albumService.DeleteMediaFromAll(mediaId)
	if svcErr != nil {
		return svcErr
	}

	err = s.mediaRepository.Delete(mediaId)
	if err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "unable to delete media")
	}

	// Finally, remove the media file from disk storage
	err = os.Remove(filepath.Join(mediaFolder, *media.StorageFileName))
	if err != nil {
		slog.Debug("Couldn't remove media file from storage", "error", err)
		return utils.NewServiceError(http.StatusInternalServerError, "unable to delete media")
	}
	return nil
}
