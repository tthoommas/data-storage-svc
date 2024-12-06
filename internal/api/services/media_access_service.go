package services

import (
	"data-storage-svc/internal/repository"
	"data-storage-svc/internal/utils"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MediaAccessService interface {
	// Grant access to a media for a given user
	GrantAccess(userId *primitive.ObjectID, mediaId *primitive.ObjectID) utils.ServiceError
	// Revoke all accesses to a media for all users
	RevokeAll(mediaId *primitive.ObjectID) utils.ServiceError
	// Check if a user can view a given media
	CanView(userId *primitive.ObjectID, mediaId *primitive.ObjectID) bool
}

type mediaAccessService struct {
	mediaAccessRepository repository.MediaAccessRepository
}

func NewMediaAccessService(mediaAccessRepository repository.MediaAccessRepository) mediaAccessService {
	return mediaAccessService{mediaAccessRepository}
}

func (s mediaAccessService) GrantAccess(userId *primitive.ObjectID, mediaId *primitive.ObjectID) utils.ServiceError {
	err := s.mediaAccessRepository.Create(userId, mediaId)
	if err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "couldn't grant access to media")
	}
	return nil
}

func (s mediaAccessService) RevokeAll(mediaId *primitive.ObjectID) utils.ServiceError {
	err := s.mediaAccessRepository.RemoveAll(mediaId)
	if err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "couldn't revoke media permissions")
	}
	return nil
}

func (s mediaAccessService) CanView(userId *primitive.ObjectID, mediaId *primitive.ObjectID) bool {
	access, err := s.mediaAccessRepository.Get(userId, mediaId)
	if err != nil {
		return false
	}
	return access != nil
}
