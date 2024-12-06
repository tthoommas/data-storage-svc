package services

import (
	"data-storage-svc/internal/model"
	"data-storage-svc/internal/repository"
	"data-storage-svc/internal/utils"
	"log/slog"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlbumAccessService interface {
	// Grant access to an album for a given user
	GrantAccess(userId *primitive.ObjectID, albumId *primitive.ObjectID, canEdit bool) utils.ServiceError
	// Check if a given user has permissions to view a given album
	CanViewAlbum(userId *primitive.ObjectID, albumId *primitive.ObjectID) bool
	// Check if a given user has permissions to edit a given album
	CanEditAlbum(userId *primitive.ObjectID, albumId *primitive.ObjectID) bool
	// Get all album access authorizations for a given user
	GetAllForUser(userId *primitive.ObjectID) ([]model.UserAlbumAccess, error)
	// Revoke all accesses granted to users to access or edit this album
	RevokeAllAccesses(albumId *primitive.ObjectID) error
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

func (s albumAccessService) CanViewAlbum(userId *primitive.ObjectID, albumId *primitive.ObjectID) bool {
	userAlbumAccess, err := s.albumAccessRepository.Get(userId, albumId)
	if err != nil {
		slog.Debug("can view album error", "error", err)
		return false
	}
	return userAlbumAccess != nil
}

func (s albumAccessService) CanEditAlbum(userId *primitive.ObjectID, albumId *primitive.ObjectID) bool {
	userAlbumAccess, err := s.albumAccessRepository.Get(userId, albumId)
	if err != nil {
		slog.Debug("can view album error", "error", err)
		return false
	}
	return userAlbumAccess != nil && userAlbumAccess.CanEdit
}

func (s albumAccessService) GetAllForUser(userId *primitive.ObjectID) ([]model.UserAlbumAccess, error) {
	return s.albumAccessRepository.GetAllByUser(userId)
}

func (s albumAccessService) RevokeAllAccesses(albumId *primitive.ObjectID) error {
	return s.albumAccessRepository.RemoveAllAccesses(albumId)
}
