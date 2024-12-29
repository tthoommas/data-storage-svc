package services

import (
	"crypto/rand"
	"data-storage-svc/internal/model"
	"data-storage-svc/internal/repository"
	"data-storage-svc/internal/utils"
	"encoding/base64"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SharedLinkService interface {
	// Create a new shared link for the given album id
	Create(albumId primitive.ObjectID, createdBy primitive.ObjectID, expirationDate time.Time, canEdit bool) (*model.SharedLink, utils.ServiceError)
	// List all shared link for the given album id
	List(albumId primitive.ObjectID) ([]model.SharedLink, utils.ServiceError)
	// Get a shared link id by its token
	GetByToken(token string) (*model.SharedLink, utils.ServiceError)
	// Delete a given shared link
	Delete(sharedLinkId primitive.ObjectID) utils.ServiceError
	// Update an existing shared link
	Update(sharedLinkId primitive.ObjectID, canEdit bool) utils.ServiceError
}

type sharedLinkService struct {
	// Repository dependencies
	sharedLinkRepository  repository.SharedLinkRepository
	albumAccessRepository repository.AlbumAccessRepository
	// Service dependencies
}

func NewSharedLinkService(sharedLinkRepository repository.SharedLinkRepository, albumAccessRepository repository.AlbumAccessRepository) SharedLinkService {
	return sharedLinkService{sharedLinkRepository: sharedLinkRepository, albumAccessRepository: albumAccessRepository}
}

func (s sharedLinkService) Create(albumId primitive.ObjectID, createdBy primitive.ObjectID, expirationDate time.Time, canEdit bool) (*model.SharedLink, utils.ServiceError) {
	access, err := s.albumAccessRepository.Get(&createdBy, &albumId)
	if err != nil || access == nil {
		return nil, utils.NewServiceError(http.StatusUnauthorized, "cannot create a shared link for this album")
	}
	token, err := generateRandomString(32)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't create the shared link")
	}
	newLink := model.SharedLink{AlbumId: albumId, CreatedBy: createdBy, CreatedAt: time.Now(), Token: token, ExpirationDate: expirationDate, CanEdit: canEdit}
	_, err = s.sharedLinkRepository.Create(&newLink)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't create the shared link")
	}
	return &newLink, nil
}

func generateRandomString(length int) (string, error) {
	// Calculate the number of bytes required
	byteLength := (length * 3) / 4

	// Create a byte slice
	randomBytes := make([]byte, byteLength)

	// Fill the byte slice with random data
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode to base64 and truncate to the required length
	return base64.RawURLEncoding.EncodeToString(randomBytes)[:length], nil
}

func (s sharedLinkService) List(albumId primitive.ObjectID) ([]model.SharedLink, utils.ServiceError) {
	result, err := s.sharedLinkRepository.List(&albumId)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't find shared links")
	}
	return result, nil
}

func (s sharedLinkService) GetByToken(token string) (*model.SharedLink, utils.ServiceError) {
	sharedLink, err := s.sharedLinkRepository.GetByToken(token)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusNotFound, "couldn't find the shared link")
	}
	return sharedLink, nil
}

func (s sharedLinkService) Delete(sharedLinkId primitive.ObjectID) utils.ServiceError {
	if err := s.sharedLinkRepository.Delete(sharedLinkId); err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "couldn't delete shared link")
	}
	return nil
}

func (s sharedLinkService) Update(sharedLinkId primitive.ObjectID, canEdit bool) utils.ServiceError {
	err := s.sharedLinkRepository.Update(sharedLinkId, canEdit)
	if err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "couldn't update shared link")
	}
	return nil
}
