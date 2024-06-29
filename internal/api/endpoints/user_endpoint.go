package endpoints

import (
	"data-storage-svc/internal/api/security"
	"data-storage-svc/internal/api/utils"
	"data-storage-svc/internal/database"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RegisterUserBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterUser(c *gin.Context) {
	var registerUserBody RegisterUserBody

	if err := c.BindJSON(&registerUserBody); err != nil {
		return
	}
	if !utils.IsValidEmail(registerUserBody.Email) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid email address"})
		return
	}
	if len(registerUserBody.Password) < 6 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "password should be at least 6 characters long"})
		return
	}
	if database.UserExists(&registerUserBody.Email) {
		slog.Debug("user already exits cannot create duplicate", "email", registerUserBody.Email)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
		return
	}
	hash, err := security.HashPassword(registerUserBody.Password)
	if err != nil {
		slog.Debug("Impossible to hash password", "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "couldn't register new user"})
		return
	}
	err = database.StoreNewUser(&registerUserBody.Email, &hash)
	if err != nil {
		slog.Debug("couldn't register new user in database", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "couldn't register new user"})
		return
	}
	slog.Debug("New user register", "email", registerUserBody.Email, "hash", hash)
}

type FetchJWTBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func FetchJWT(c *gin.Context) {
	var fetchJWTBody FetchJWTBody

	if err := c.BindJSON(&fetchJWTBody); err != nil {
		return
	}
	if !security.AuthenticateUser(&fetchJWTBody.Email, &fetchJWTBody.Password) {
		slog.Debug("Auth failed for user", "user", fetchJWTBody.Email)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid login"})
		return
	}

	jwt, err := security.CreateToken(&fetchJWTBody.Email)
	if err != nil {
		slog.Debug("Couldn't create JWT", "email", fetchJWTBody.Email, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "couldn't create token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"jwt": jwt})
}
