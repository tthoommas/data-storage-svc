package services_test

import (
	"data-storage-svc/internal/api/services"
	"data-storage-svc/internal/mocks"
	"data-storage-svc/internal/model"
	"data-storage-svc/internal/utils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestCreateUser(t *testing.T) {
	testCases := []struct {
		name        string
		email       string
		password    string
		expectError bool
	}{
		{
			name:        "Invalid email",
			email:       "invalidEmail",
			password:    "password",
			expectError: true,
		},
		{
			name:        "To short password",
			email:       "test@test.fr",
			password:    "a",
			expectError: true,
		},
		{
			name:        "User already exists",
			email:       "alreadyexists@test.fr",
			password:    "aksopkapoksop",
			expectError: true,
		},
		{
			name:        "Valid registration",
			email:       "test@test.fr",
			password:    "aVeryRandomPassword@~°12",
			expectError: false,
		},
	}

	mockRepository := &mocks.UserRepository{}
	hashModule := &mocks.HashModule{}
	tokenModule := &mocks.TokenModule{}
	svc := services.NewUserService(mockRepository, hashModule, tokenModule)

	newObjId := primitive.NewObjectID()
	mockRepository.On("Create", mock.Anything).Return(&newObjId, nil)
	mockRepository.On("GetByEmail", "alreadyexists@test.fr").Return(&model.User{}, nil)
	mockRepository.On("GetByEmail", "test@test.fr").Return((*model.User)(nil), mongo.ErrNoDocuments)

	hashModule.On("HashPassword", "aVeryRandomPassword@~°12").Return("hashedpassword", nil)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createdId, err := svc.Create(tc.email, tc.password)
			if tc.expectError {
				assert.Nil(t, createdId)
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, createdId)
			}
			for _, c := range mockRepository.Calls {
				if c.Method == "Create" {
					user := c.Arguments.Get(0).(*model.User)
					// Assert we do not register clear password in db
					assert.Equal(t, "hashedpassword", user.PasswordHash)
				}
			}
		})
	}
}

func TestGetByEmail(t *testing.T) {
	testCases := []struct {
		name            string
		email           string
		expectErrorCode *int
	}{
		{
			name:            "Unexisting user",
			email:           "unexisting@test.fr",
			expectErrorCode: utils.IntPtr(404),
		},
		{
			name:            "Existing user",
			email:           "test@test.fr",
			expectErrorCode: nil,
		},
	}

	mockRepository := &mocks.UserRepository{}
	mockRepository.On("GetByEmail", "unexisting@test.fr").Return((*model.User)(nil), mongo.ErrNoDocuments)
	mockRepository.On("GetByEmail", "test@test.fr").Return(&model.User{Email: "test@test.fr"}, nil)

	svc := services.NewUserService(mockRepository, nil, nil)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := svc.GetByEmail(tc.email)
			if tc.expectErrorCode != nil {
				assert.NotNil(t, err)
				assert.Equal(t, *tc.expectErrorCode, err.GetCode())
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, user)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	testCases := []struct {
		name            string
		email           string
		password        string
		expectErrorCode *int
	}{
		{
			name:            "Unexisting user",
			email:           "unexisting@test.fr",
			password:        "azertyuiop",
			expectErrorCode: utils.IntPtr(401),
		},
		{
			name:            "Invalid password",
			email:           "test@test.fr",
			password:        "invalidpassword",
			expectErrorCode: utils.IntPtr(401),
		},
		{
			name:            "Valid user/Valid password",
			email:           "test@test.fr",
			password:        "azertyuiop",
			expectErrorCode: nil,
		},
	}
	hash := "$2a$14$RUahhb6.L8oVMq91f3.HQ.37SKrtcmAkFwp8lW.eb7WFJy9G6ZayK"

	mockRepository := &mocks.UserRepository{}
	mockRepository.On("GetByEmail", "unexisting@test.fr").Return((*model.User)(nil), mongo.ErrNoDocuments)
	userId := primitive.NewObjectID()
	mockRepository.On("GetByEmail", "test@test.fr").Return(&model.User{Id: userId, Email: "test@test.fr", PasswordHash: hash}, nil)
	mockRepository.On("Update", &userId, mock.Anything).Return(nil)

	hashModule := &mocks.HashModule{}
	hashModule.On("VerifyPassword", "invalidpassword", hash).Return(false)
	hashModule.On("VerifyPassword", "azertyuiop", hash).Return(true)

	tokenModule := &mocks.TokenModule{}
	tokenModule.On("CreateToken", utils.StrPtr("test@test.fr")).Return("asecretgeneratedtoken", nil)

	svc := services.NewUserService(mockRepository, hashModule, tokenModule)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := svc.GenerateToken(tc.email, tc.password)
			if tc.expectErrorCode != nil {
				assert.NotNil(t, err)
				assert.Equal(t, *tc.expectErrorCode, err.GetCode())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, "asecretgeneratedtoken", *token)
			}
		})
	}
}
