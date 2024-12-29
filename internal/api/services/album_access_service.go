package services

import (
	"data-storage-svc/internal/model"
	"data-storage-svc/internal/repository"
	"data-storage-svc/internal/utils"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlbumAccessService interface {
	// Grant access to an album for a given user
	GrantAccess(userId *primitive.ObjectID, albumId *primitive.ObjectID, canEdit bool) utils.ServiceError
	// Get all album access authorizations for a given user
	GetAllForUser(userId *primitive.ObjectID) ([]model.UserAlbumAccess, error)
	// Revoke access for the given user and given album
	RevokeAccess(userId *primitive.ObjectID, albumId *primitive.ObjectID) utils.ServiceError
	// Revoke all accesses granted to users to access or edit this album
	RevokeAllAccesses(albumId *primitive.ObjectID) error
	// List all accesses granted for a given album
	GetAllAccesses(albumId *primitive.ObjectID) ([]model.UserAlbumAccess, utils.ServiceError)
}

type albumAccessService struct {
	albumAccessRepository repository.AlbumAccessRepository
}

func NewAlbumAccessService(albumAccessRepository repository.AlbumAccessRepository) AlbumAccessService {
	return albumAccessService{albumAccessRepository}
}

func (s albumAccessService) GrantAccess(userId *primitive.ObjectID, albumId *primitive.ObjectID, canEdit bool) utils.ServiceError {
	err := s.albumAccessRepository.Create(userId, albumId, canEdit)
	if err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "couldn't grant access to album")
	}
	return nil
}

func (s albumAccessService) GetAllForUser(userId *primitive.ObjectID) ([]model.UserAlbumAccess, error) {
	return s.albumAccessRepository.GetAllByUser(userId)
}

func (s albumAccessService) RevokeAccess(userId *primitive.ObjectID, albumId *primitive.ObjectID) utils.ServiceError {
	if err := s.albumAccessRepository.Remove(userId, albumId); err != nil {
		return utils.NewServiceError(http.StatusNotFound, "couldn't revoke access")
	}
	return nil
}

func (s albumAccessService) RevokeAllAccesses(albumId *primitive.ObjectID) error {
	return s.albumAccessRepository.RemoveAllAccesses(albumId)
}

func (s albumAccessService) GetAllAccesses(albumId *primitive.ObjectID) ([]model.UserAlbumAccess, utils.ServiceError) {
	accesses, err := s.albumAccessRepository.GetAllByAlbum(albumId)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't find album accesses")
	}
	return accesses, nil
}
