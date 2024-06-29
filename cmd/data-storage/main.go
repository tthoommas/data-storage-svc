package main

import (
	"data-storage-svc/internal/api/endpoints"
	"data-storage-svc/internal/api/middlewares"
	"data-storage-svc/internal/deployment"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	slog.Info("Starting data-storage service...")
	deployment.StartMongoDB()
	router := gin.Default()
	router.POST("/registerUser", endpoints.RegisterUser)
	router.POST("/fetchJwt", endpoints.FetchJWT)

	authorized := router.Group("", middlewares.AuthMiddleware())
	{
		authorized.POST("/upload", endpoints.UploadMedia)
	}
	router.Run("0.0.0.0:8080")
}
