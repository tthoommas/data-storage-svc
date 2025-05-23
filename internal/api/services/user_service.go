package services

import (
	"data-storage-svc/internal/api/security"
	"data-storage-svc/internal/model"
	"data-storage-svc/internal/repository"
	"data-storage-svc/internal/utils"
	"log/slog"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserService interface {
	// Create a new user (registration)
	Create(email string, password string) (*primitive.ObjectID, utils.ServiceError)
	// Get by email
	GetByEmail(email string) (*model.User, utils.ServiceError)
	// Get by id
	GetById(userId primitive.ObjectID) (*model.User, utils.ServiceError)
	// Get all
	GetAll() ([]model.User, utils.ServiceError)
	// Generates a JWT token for the given user
	GenerateToken(email string, password string) (*string, utils.ServiceError)
}

type userService struct {
	// Repository dependencies
	userRepository repository.UserRepository
	hashModule     security.HashModule
	tokenModule    security.TokenModule
}

func NewUserService(userRepository repository.UserRepository, hashModule security.HashModule, tokenModule security.TokenModule) userService {
	return userService{userRepository, hashModule, tokenModule}
}

func (s userService) Create(email string, password string) (*primitive.ObjectID, utils.ServiceError) {
	// Check email
	if !utils.IsValidEmail(email) {
		return nil, utils.NewServiceError(http.StatusBadRequest, "invalid email address")
	}
	// Check password
	if len(password) < 6 {
		return nil, utils.NewServiceError(http.StatusBadRequest, "password should be at least 6 characters long")
	}
	// Check if user already exists
	_, err := s.userRepository.GetByEmail(email)
	if err == nil || err != mongo.ErrNoDocuments {
		return nil, utils.NewServiceError(http.StatusBadRequest, "couldn't create new user")
	}
	// Hash the user's password
	hash, err := s.hashModule.HashPassword(password)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't create new user")
	}
	// Finally store the user in DB
	createdId, err := s.userRepository.Create(&model.User{Email: email, PasswordHash: hash, IsAdmin: false, JoinDate: time.Now()})
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't create new user")
	}
	return createdId, nil
}

func (s userService) GetByEmail(email string) (*model.User, utils.ServiceError) {
	user, err := s.userRepository.GetByEmail(email)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusNotFound, "unable to find user")
	}
	return user, nil
}

func (s userService) GenerateToken(email string, password string) (*string, utils.ServiceError) {
	if len(email) == 0 || len(password) == 0 {
		return nil, utils.NewServiceError(http.StatusUnauthorized, "unable to authenticate user")
	}

	user, err := s.userRepository.GetByEmail(email)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusUnauthorized, "unable to authenticate user")
	}
	if !s.hashModule.VerifyPassword(password, user.PasswordHash) {
		return nil, utils.NewServiceError(http.StatusUnauthorized, "unable to authenticate user")
	}
	// User is authenticated, generate a new token
	jwt, err := s.tokenModule.CreateToken(&user.Email)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't generate token")
	}
	// Update the last login date
	if s.userRepository.Update(&user.Id, bson.M{"$set": bson.M{"lastLogin": time.Now()}}) != nil {
		slog.Error("couldn't update join date for user", "userId", user.Id.Hex())
	}
	return &jwt, nil
}

func (s userService) GetById(userId primitive.ObjectID) (*model.User, utils.ServiceError) {
	user, err := s.userRepository.GetById(&userId)
	if err != nil {
		return nil, utils.NewServiceError(http.StatusNotFound, "couldn't find user")
	}
	return user, nil
}

func (s userService) GetAll() ([]model.User, utils.ServiceError) {
	users, err := s.userRepository.GetAll()
	if err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "couldn't fetch user list")
	}
	return users, nil
}
