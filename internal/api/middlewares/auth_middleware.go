package middlewares

import (
	"data-storage-svc/internal/api/security"
	"data-storage-svc/internal/api/services"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AuthMiddleware(userService services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authCookie, err := c.Cookie("jwt")
		if err != nil || authCookie == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		tokenString := authCookie
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return security.GetSecretKey(), nil
		})

		if err != nil || !token.Valid {
			slog.Debug("Couldn't validate JWT", "error", err, "tokenValidity", token.Valid)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Extract email claim
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			email := claims["email"].(string)
			user, err := userService.GetByEmail(email)
			if err != nil {
				slog.Debug("couldn't find authenticated user", "email", email, "error", err)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				return
			}
			c.Set("user", user)
		} else {
			slog.Debug("Couldn't find email claim in token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		c.Next()
	}
}
