package services

import (
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/model"
	"data-storage-svc/internal/repository"
	"data-storage-svc/internal/utils"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/evanoberholster/imagemeta"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MediaService interface {
	// Create a new media resource
	Create(fileName string, uploader *primitive.ObjectID, uploadedViaSharedLink bool, data *io.ReadCloser) (*primitive.ObjectID, utils.ServiceError)
	// Get media by id
	GetById(mediaId *primitive.ObjectID) (*model.Media, utils.ServiceError)
	// Get the media data (i.e. bytes of the file stored on disk)
	GetData(storageFileName string, compressed bool) (*string, []byte, utils.ServiceError)
	// Get the media metadata (i.e. exif data contained in original file)
	GetMetaData(mediaId *primitive.ObjectID) (*model.MetaData, utils.ServiceError)
	// Get all media accessible to a given user
	GetAllSharedWithUser(userId *primitive.ObjectID) ([]model.UserMediaAccess, utils.ServiceError)
	// Get all medias uploaded by user
	GetAllUploadedByUser(userId *primitive.ObjectID) ([]model.Media, utils.ServiceError)
	// Delete a specific media
	Delete(mediaId *primitive.ObjectID) utils.ServiceError
	// Check if a media is in a given album
	IsInAlbum(mediaId *primitive.ObjectID, albumId *primitive.ObjectID) bool
}

type mediaService struct {
	// Repository dependencies
	mediaRepository        repository.MediaRepository
	mediaInAblumRepository repository.MediaInAlbumRepository
	// Service dependencies
	mediaAccessService MediaAccessService
	albumService       AlbumService
}

func NewMediaService(mediaRepository repository.MediaRepository, mediaInAblumRepository repository.MediaInAlbumRepository, mediaAccessService MediaAccessService, albumService AlbumService) mediaService {
	return mediaService{mediaRepository, mediaInAblumRepository, mediaAccessService, albumService}
}

func (s mediaService) Create(fileName string, uploader *primitive.ObjectID, uploadedViaSharedLink bool, data *io.ReadCloser) (*primitive.ObjectID, utils.ServiceError) {
	if len(fileName) == 0 {
		return nil, utils.NewServiceError(http.StatusBadRequest, "no file name provided while uploading")
	}
	if data == nil {
		return nil, utils.NewServiceError(http.StatusBadRequest, "no file provided")
	}
	extension, isValid := checkExtension(&fileName)
	if !isValid {
		return nil, utils.NewServiceError(http.StatusBadRequest, "invalid filename")
	}
	targetFolderOriginal, err := utils.GetDataDir("originalMedias")
	targetFolderCompressed, err2 := utils.GetDataDir("compressedMedias")
	if err != nil || err2 != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't upload file")
	}
	storageFileName := uuid.NewString()
	storageFileName = storageFileName + "." + extension
	uploadTime := time.Now()
	mediaId, err := s.mediaRepository.Create(
		&model.Media{
			OriginalFileName:      &fileName,
			StorageFileName:       &storageFileName,
			UploadedBy:            uploader,
			UploadTime:            &uploadTime,
			UploadedViaSharedLink: uploadedViaSharedLink,
		},
	)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusBadRequest, "couldn't upload file")
	}
	// Store original file on disk
	targetFileOriginal := filepath.Join(targetFolderOriginal, storageFileName)
	outFile, err := os.Create(targetFileOriginal)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't upload file")
	}
	defer outFile.Close()
	defer (*data).Close()

	_, err = io.Copy(outFile, *data)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "upload failed")
	}

	// Store JPEG compressed version on disk
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "upload failed")
	}
	original, err := bimg.Read(targetFileOriginal)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "upload failed")
	}
	jpgImage, err := bimg.NewImage(original).Convert(bimg.JPEG)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "upload failed")
	}
	compressedJpgImage, err := bimg.NewImage(jpgImage).Process(bimg.Options{Quality: 20, NoAutoRotate: true, StripMetadata: true})
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "upload failed")
	}
	targetFileCompressed := filepath.Join(targetFolderCompressed, storageFileName)
	if bimg.Write(targetFileCompressed, compressedJpgImage) != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "upload failed")
	}
	return mediaId, nil
}

var MediaExtensionMimeType = map[string]string{
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"gif":  "image/gif",
	"heic": "image/heic",
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
	return strings.ToLower(extension), nil
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

func readMediaFromStorage(directory, storageFileName *string) (*string, []byte, error) {
	if directory == nil || storageFileName == nil || len(*directory) == 0 || len(*storageFileName) == 0 {
		return nil, nil, fmt.Errorf("couldn't read file with empty directory/filename")
	}

	mediaFilePath := filepath.Join(*directory, *storageFileName)
	slog.Debug("Reading requested file", "filePath", mediaFilePath)
	data, err := os.ReadFile(mediaFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't find media")
	}

	// Infer mime type and send the media
	mimeType, err := getMimeType(storageFileName)
	if err != nil {
		slog.Debug("Cannot infer file mime type", "error", err, "mediaFilePath", mediaFilePath)
		return nil, nil, fmt.Errorf("couldn't find media")
	}
	return &mimeType, data, nil
}

func (s mediaService) GetById(mediaId *primitive.ObjectID) (*model.Media, utils.ServiceError) {
	media, err := s.mediaRepository.Get(mediaId)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusNotFound, "media not found")
	}
	return media, nil
}

func (s mediaService) GetData(storageFileName string, compressed bool) (*string, []byte, utils.ServiceError) {
	// Read media to send it
	directory := ""
	if compressed {
		directory = common.COMPRESSED_DIRECTORY
	} else {
		directory = common.ORIGINAL_MEDIA_DIRECTORY
	}
	mediaDirectory, err := utils.GetDataDir(directory)
	if err != nil {
		return nil, nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't find media")
	}

	mimeType, data, err := readMediaFromStorage(&mediaDirectory, &storageFileName)
	if err != nil {
		return nil, nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't find the requested media")
	}
	return mimeType, data, nil
}

func (s mediaService) GetMetaData(mediaId *primitive.ObjectID) (*model.MetaData, utils.ServiceError) {
	media, svcErr := s.GetById(mediaId)
	if svcErr != nil {
		return nil, svcErr
	}
	mediaDir, err := utils.GetDataDir("originalMedias")
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't get meta data")
	}
	mediaFilePath := filepath.Join(mediaDir, *media.StorageFileName)
	mediaFile, err := os.Open(mediaFilePath)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't get meta data")
	}
	defer mediaFile.Close()

	exif, err := imagemeta.Decode(mediaFile)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusNotFound, "no meta data found for this image")
	}

	metaData := model.MetaData{
		CameraModel: exif.LensModel,
		Location: &model.Location{
			Latitude:  exif.GPS.Latitude(),
			Longitude: exif.GPS.Longitude(),
			Altitude:  exif.GPS.Altitude(),
		},
		Created: exif.CreateDate(),
	}
	return &metaData, nil
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
	originalMediaFolder, err := utils.GetDataDir("originalMedias")
	compressedMediaFolder, err2 := utils.GetDataDir("compressedMedias")

	if err != nil || err2 != nil {
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
	err = os.Remove(filepath.Join(originalMediaFolder, *media.StorageFileName))
	err2 = os.Remove(filepath.Join(compressedMediaFolder, *media.StorageFileName))
	if err != nil || err2 != nil {
		slog.Debug("Couldn't remove media file from storage", "error", err)
		return utils.NewServiceError(http.StatusInternalServerError, "unable to delete media")
	}
	return nil
}

func (s mediaService) IsInAlbum(mediaId *primitive.ObjectID, albumId *primitive.ObjectID) bool {
	return s.mediaInAblumRepository.IsInAlbum(mediaId, albumId)
}
