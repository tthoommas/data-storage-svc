package middlewares

import (
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/api/security"
	"data-storage-svc/internal/api/services"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AuthMiddleware(userService services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If there is a shared link, user do not enforce user authentication
		_, hasSharedLink := c.Get(common.SHARED_LINK)
		authCookie, err := c.Cookie("jwt")
		if err != nil || authCookie == "" {
			if !hasSharedLink {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
				return
			} else {
				// Shared link present, do not reject the request now
				c.Next()
				return
			}
		}

		tokenString := authCookie
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return security.GetSecretKey(), nil
		})

		if err != nil || !token.Valid {
			if !hasSharedLink {
				slog.Debug("Couldn't validate JWT", "error", err, "tokenValidity", token.Valid)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				return
			} else {
				// Shared link is present
				c.Next()
				return
			}
		}

		// Extract email claim
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			email := claims["email"].(string)
			user, err := userService.GetByEmail(email)
			if err != nil {
				if !hasSharedLink {
					slog.Debug("couldn't find authenticated user", "email", email, "error", err)
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
					return
				} else {
					c.Next()
					return
				}
			}
			c.Set(common.USER, user)
		} else {
			if !hasSharedLink {
				slog.Debug("Couldn't find email claim in token")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
				return
			}
		}

		c.Next()
	}
}
