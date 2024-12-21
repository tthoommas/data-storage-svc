package services

import (
	"archive/zip"
	"data-storage-svc/internal/model"
	"data-storage-svc/internal/repository"
	"data-storage-svc/internal/utils"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DownloadService interface {
	// Initiate the creation of a zip file (to download the album)
	InitDownload(albumId *primitive.ObjectID, initiator *primitive.ObjectID) (*primitive.ObjectID, utils.ServiceError)
	// Check if a download is ready to be downloaded
	IsReady(downloadId *primitive.ObjectID) bool
	// Get a download by id
	Get(downloadId *primitive.ObjectID) (*model.Download, utils.ServiceError)
	// Get the data of the download, ie. the zip file bytes
	GetData(downloadId *primitive.ObjectID) ([]byte, utils.ServiceError)
}

type downloadService struct {
	// Repository dependencies
	albumRepository        repository.AlbumRepository
	downloadRepository     repository.DownloadRepository
	mediaRepository        repository.MediaRepository
	mediaInAlbumRepository repository.MediaInAlbumRepository
	// Service dependencies
}

func NewDownloadService(albumRepository repository.AlbumRepository, downloadRepository repository.DownloadRepository, mediaRepository repository.MediaRepository, mediaInAlbumRepository repository.MediaInAlbumRepository) downloadService {
	return downloadService{albumRepository: albumRepository, downloadRepository: downloadRepository, mediaRepository: mediaRepository, mediaInAlbumRepository: mediaInAlbumRepository}
}

func (s downloadService) InitDownload(albumId *primitive.ObjectID, initiator *primitive.ObjectID) (*primitive.ObjectID, utils.ServiceError) {

	album, err := s.albumRepository.GetById(*albumId)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusNotFound, "album not found")
	}

	mediasInAlbum, err := s.mediaInAlbumRepository.ListAllMedias(albumId)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusBadRequest, "couldn't find album or medias")
	}

	// Retrieve all medias to download
	var mediasToDownload []model.Media = make([]model.Media, 0)
	for _, mediaInAlbum := range mediasInAlbum {
		media, err := s.mediaRepository.Get(mediaInAlbum.MediaId)
		if err == nil {
			mediasToDownload = append(mediasToDownload, *media)
		} else {
			slog.Error("unable to find media", "mediaId", mediaInAlbum.MediaId.String())
		}
	}

	// Create a new zip archive file name in DB
	fileId := uuid.New()
	zipFileName := fileId.String() + ".zip"
	now := time.Now()
	downloadId, err := s.downloadRepository.Create(&model.Download{DownloadName: album.Title, StartedAt: &now, ZipFileName: &zipFileName, IsReady: false, Initiator: initiator})
	if err != nil || downloadId == nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't initiate download")
	}

	// Start the zip archive file creation, asynchronously
	go s.createZipFile(zipFileName, *downloadId, mediasToDownload)

	return downloadId, nil
}

// Create the zip file to download. Must be called in a different go routine as it can take a very long time
func (s downloadService) createZipFile(zipFileName string, downloadId primitive.ObjectID, medias []model.Media) error {
	downloadFolder, err := utils.GetDataDir("downloads")
	if err != nil {
		slog.Debug("couldn't open media folder", "error", err)
		return err
	}
	fileLocation := filepath.Join(downloadFolder, zipFileName)
	slog.Debug("Creating a new zip archive", "zipFileLocation", fileLocation)
	zipFileArchive, err := os.Create(fileLocation)
	if err != nil {
		panic(err)
	}
	defer zipFileArchive.Close()

	mediaFolder, err := utils.GetDataDir("medias")
	if err != nil {
		return fmt.Errorf("couldn't create zip archive, cannot find medias folder")
	}

	zipWriter := zip.NewWriter(zipFileArchive)
	// Don't forget to close the zip writer
	defer zipWriter.Close()

	// Write all media files inside the zip file
	for _, media := range medias {
		// Open the media file to be written in zip file
		mediaFile, err := os.Open(filepath.Join(mediaFolder, *media.StorageFileName))
		if err == nil {
			// Create a writer targetting the zip file
			writer, err := zipWriter.Create(*media.OriginalFileName)
			if err != nil {
				slog.Error("Couldn't open zip file writer for file", "err", err)
			} else {
				// Writing the file into the zip file
				_, err := io.Copy(writer, mediaFile)
				if err != nil {
					slog.Error("couldn't write file in zip file", "error", err)
				}
			}
			// Ultimately, close the original file
			mediaFile.Close()
		} else {
			slog.Error("Couldn't open media file for download", "mediaFile", media.Id.String())
		}
	}
	s.downloadRepository.MarkAsReady(&downloadId)
	slog.Debug("Zip file created", "zipFileLocation", fileLocation)
	return nil
}

func (s downloadService) IsReady(downloadId *primitive.ObjectID) bool {
	download, err := s.downloadRepository.Get(downloadId)
	if err != nil {
		return false
	}
	return download.IsReady
}

func (s downloadService) Get(downloadId *primitive.ObjectID) (*model.Download, utils.ServiceError) {
	download, err := s.downloadRepository.Get(downloadId)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusNotFound, "couldn't find the download")
	}
	return download, nil
}

func (s downloadService) GetData(downloadId *primitive.ObjectID) ([]byte, utils.ServiceError) {
	downloadDir, err := utils.GetDataDir("downloads")
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't find download")
	}
	download, svcErr := s.Get(downloadId)
	if svcErr != nil {
		return nil, svcErr
	}
	// Check that the download is ready before downloading (otherwise it will be corrupted)
	if !download.IsReady {
		return nil, utils.NewServiceError(http.StatusBadRequest, "download is not yet ready")
	}

	data, err := os.ReadFile(filepath.Join(downloadDir, *download.ZipFileName))
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "download error")
	}
	return data, nil
}
