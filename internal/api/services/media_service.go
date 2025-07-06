package services

import (
	"bytes"
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/model"
	"data-storage-svc/internal/repository"
	"data-storage-svc/internal/utils"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/evanoberholster/imagemeta"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MediaService interface {
	// Create a new media resource
	Create(fileName string, uploader *primitive.ObjectID, uploadedViaSharedLink bool, data io.ReadCloser) (*primitive.ObjectID, utils.ServiceError)
	// Get media by id
	GetById(mediaId *primitive.ObjectID) (*model.Media, utils.ServiceError)
	// Get the media data (i.e. bytes of the file stored on disk)
	GetData(storageFileName string, compressed bool) (*os.File, *time.Time, utils.ServiceError)
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

func (s mediaService) Create(fileName string, uploader *primitive.ObjectID, uploadedViaSharedLink bool, data io.ReadCloser) (*primitive.ObjectID, utils.ServiceError) {
	if len(fileName) == 0 {
		return nil, utils.NewServiceError(http.StatusBadRequest, "no file name provided while uploading")
	}
	if data == nil {
		return nil, utils.NewServiceError(http.StatusBadRequest, "could't create requested file, no content provided")
	}
	fileHeader := make([]byte, 512)
	n, err := data.Read(fileHeader)
	if err != nil || n != 512 {
		return nil, utils.NewServiceError(http.StatusBadRequest, "couldn't read file header")
	}
	_, extension, isValid := utils.CheckFileExtension(fileHeader)
	if !isValid {
		return nil, utils.NewServiceError(http.StatusBadRequest, "invalid file format")
	}
	targetFolderOriginal, err := utils.GetDataDir("originalMedias")
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't upload file")
	}
	storageUUID := uuid.NewString()
	originalStorageFileName := storageUUID + "." + extension
	uploadTime := time.Now()
	mediaId, err := s.mediaRepository.Create(
		&model.Media{
			OriginalFileName:      &fileName,
			StorageFileName:       &originalStorageFileName,
			UploadedBy:            uploader,
			UploadTime:            &uploadTime,
			UploadedViaSharedLink: uploadedViaSharedLink,
		},
	)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusBadRequest, "couldn't upload file")
	}
	// Store original file on disk
	targetFileOriginal := filepath.Join(targetFolderOriginal, originalStorageFileName)
	outFile, err := os.Create(targetFileOriginal)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't upload file")
	}

	multi := io.MultiReader(bytes.NewReader(fileHeader), data)
	_, err = io.Copy(outFile, multi)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "upload failed")
	}

	return mediaId, nil
}

func (s mediaService) GetById(mediaId *primitive.ObjectID) (*model.Media, utils.ServiceError) {
	media, err := s.mediaRepository.Get(mediaId)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusNotFound, "media not found")
	}
	return media, nil
}

func (s mediaService) GetData(storageFileName string, compressed bool) (*os.File, *time.Time, utils.ServiceError) {
	// Choose the compressed version if needed
	directory := ""
	if compressed {
		directory = common.COMPRESSED_DIRECTORY
	} else {
		directory = common.ORIGINAL_MEDIA_DIRECTORY
	}
	// Use corresponding directory
	mediaDirectory, err := utils.GetDataDir(directory)
	if err != nil {
		return nil, nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't find media")
	}
	// Open file
	file, err := os.Open(filepath.Join(mediaDirectory, storageFileName))
	if err != nil {
		return nil, nil, utils.NewServiceError(http.StatusNotFound, "couldn't open requested file")
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't read file info")
	}
	modTime := fileInfo.ModTime()
	return file, &modTime, nil
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
