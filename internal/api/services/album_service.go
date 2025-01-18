package services

import (
	"data-storage-svc/internal/model"
	"data-storage-svc/internal/repository"
	"data-storage-svc/internal/utils"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlbumService interface {
	// Create an album
	Create(album *model.Album) (*primitive.ObjectID, utils.ServiceError)
	// Get album by id
	GetAlbumById(albumId *primitive.ObjectID) (*model.Album, utils.ServiceError)
	// Get all albums accessibles for a given user
	GetAllAlbumsForUser(userId *primitive.ObjectID) ([]model.Album, utils.ServiceError)
	// Get all medias in a given album
	GetMedias(albumId *primitive.ObjectID) ([]model.MediaInAlbum, utils.ServiceError)
	// Add a media to the given album
	AddMedia(mediaInAlbum *model.MediaInAlbum) utils.ServiceError
	// Delete a media from an album
	DeleteMedia(mediaId *primitive.ObjectID, albumId *primitive.ObjectID) utils.ServiceError
	// Delete a media from all albums
	DeleteMediaFromAll(mediaId *primitive.ObjectID) utils.ServiceError
	// Delete an album
	Delete(albumId *primitive.ObjectID) utils.ServiceError
}

type albumService struct {
	// Repository dependencies
	albumRepository        repository.AlbumRepository
	mediaInAlbumRepository repository.MediaInAlbumRepository
	sharedLinkRepository   repository.SharedLinkRepository

	// Service dependencies
	albumAccessService AlbumAccessService
}

func NewAlbumService(albumRepository repository.AlbumRepository, mediaInAlbumRepository repository.MediaInAlbumRepository, albumAccessService AlbumAccessService, sharedLinkRepository repository.SharedLinkRepository) albumService {
	return albumService{albumRepository, mediaInAlbumRepository, sharedLinkRepository, albumAccessService}
}

func (s albumService) Create(album *model.Album) (*primitive.ObjectID, utils.ServiceError) {
	if len(album.Title) == 0 {
		return nil, utils.NewServiceError(http.StatusBadRequest, "Cannot create album with an empty title")
	}
	// Actually create the album
	albumId, err := s.albumRepository.Create(album)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't create album")
	}
	// Grant edit access to the owner
	svcErr := s.albumAccessService.GrantAccess(album.AuthorId, albumId, true)
	if svcErr != nil {
		return nil, svcErr
	}
	return albumId, nil
}

func (s albumService) GetAlbumById(albumId *primitive.ObjectID) (*model.Album, utils.ServiceError) {
	album, err := s.albumRepository.GetById(*albumId)

	if err != nil {
		return nil, utils.NewServiceError(http.StatusNotFound, "Album not found")
	}
	return album, nil
}

func (s albumService) GetAllAlbumsForUser(userId *primitive.ObjectID) ([]model.Album, utils.ServiceError) {
	albumAccesses, err := s.albumAccessService.GetAllForUser(userId)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusNotFound, "No album found")
	}
	var albums []model.Album = make([]model.Album, 0)
	for _, access := range albumAccesses {
		album, err := s.GetAlbumById(access.AlbumId)
		if err != nil {
			return nil, err
		}
		albums = append(albums, *album)
	}
	return albums, nil
}

func (s albumService) GetMedias(albumId *primitive.ObjectID) ([]model.MediaInAlbum, utils.ServiceError) {
	medias, err := s.mediaInAlbumRepository.ListAllMedias(albumId)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusNotFound, "no medias found for this album")
	}
	return medias, nil
}

func (s albumService) AddMedia(mediaInAlbum *model.MediaInAlbum) utils.ServiceError {
	err := s.mediaInAlbumRepository.AddMediaToAlbum(mediaInAlbum)
	if err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "unable to add media to album")
	}
	return nil
}

func (s albumService) DeleteMedia(mediaId *primitive.ObjectID, albumId *primitive.ObjectID) utils.ServiceError {
	err := s.mediaInAlbumRepository.RemoveMediaFromAlbum(albumId, mediaId)
	if err != nil {
		return utils.NewServiceError(http.StatusNotFound, "unbale to remove media from the album")
	}
	return nil
}

func (s albumService) DeleteMediaFromAll(mediaId *primitive.ObjectID) utils.ServiceError {
	err := s.mediaInAlbumRepository.RemoveMediaFromAllAlbums(mediaId)
	if err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "couldn't delete media from albums")
	}
	return nil
}

func (s albumService) Delete(albumId *primitive.ObjectID) utils.ServiceError {
	// First, remove all medias from the album
	err := s.mediaInAlbumRepository.UnlinkAlbumFromAllMedias(albumId)
	if err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "couldn't delete album")
	}
	// Then remove all authorization to access this album
	err = s.albumAccessService.RevokeAllAccesses(albumId)
	if err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "couldn't delete album")
	}
	// Remove all shared links pointing to this album
	sharedLinks, err := s.sharedLinkRepository.List(albumId)
	if err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "couldn't delete album")
	}
	for _, sharedLink := range sharedLinks {
		s.sharedLinkRepository.Delete(sharedLink.Id)
	}
	// Finally remove the album
	err = s.albumRepository.Delete(albumId)
	if err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "couldn't delete album")
	}
	return nil
}
