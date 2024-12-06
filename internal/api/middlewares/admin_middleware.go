package middlewares

import (
	"data-storage-svc/internal/model"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			slog.Debug("Admin middleware: unable to get logged user")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
			return
		}
		loggedUser := user.(*model.User)
		if loggedUser == nil {
			slog.Debug("No authenticated user")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
			return
		}
		if !loggedUser.IsAdmin {
			slog.Debug("Admin middleware: user is not admin, rejected.", "email", loggedUser.Email, "isAdmin", loggedUser.IsAdmin)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
			return
		}
		slog.Debug("Admin user, allowed.", "email", loggedUser.Email)
		c.Next()
	}
}
