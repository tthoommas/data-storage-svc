package security

import (
	"data-storage-svc/internal/utils"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

var secretKey []byte = nil

func CreateToken(Email *string) (string, error) {
	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"email": Email,
			"exp":   time.Now().Add(time.Hour * 24).Unix(),
		})

	// Sign the token
	tokenString, err := token.SignedString(GetSecretKey())
	if err != nil {
		slog.Debug("Couldn't sign the JWT", "err", err)
		return "", err
	}
	return tokenString, nil
}

func GetSecretKey() []byte {
	if secretKey == nil {
		loadedKey, err := loadSecretKey()
		if err != nil {
			slog.Error("Couldn't load the JWT secret key", "err", err)
			panic(err)
		}
		secretKey = loadedKey
	}
	return secretKey
}

func loadSecretKey() ([]byte, error) {
	secretKeyFilePath, err := utils.GetDataDir("security/jwt_key")
	if err != nil {
		slog.Debug("Cannot load secret key")
		return nil, err
	}

	loadedKey, err := os.ReadFile(secretKeyFilePath)
	if err != nil {
		slog.Debug("Cannot read secretKey file content")
		return nil, err
	}
	slog.Debug("JWT secret key successfully loaded")
	return loadedKey, nil
}
