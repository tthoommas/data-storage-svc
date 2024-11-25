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
	router.Use(middlewares.CORSMiddleware())

	router.POST("/registerUser", endpoints.RegisterUser)
	router.POST("/fetchJwt", endpoints.FetchJWT)

	authorized := router.Group("", middlewares.AuthMiddleware())
	{
		authorized.POST("/upload", endpoints.UploadMedia)
		authorized.GET("/myMedias", endpoints.MyMedias)
		authorized.GET("/media", endpoints.GetMedia)
		authorized.POST("/createAlbum", endpoints.CreateAlbum)
		authorized.POST("/addMediaToAlbum", endpoints.AddMediaToAlbum)
		authorized.GET("/myAlbums", endpoints.GetMyAlbums)
		authorized.GET("/mediasInAlbum", endpoints.GetMediasInAlbum)
		authorized.GET("/album", endpoints.GetAlbumById)
		authorized.DELETE("/deleteAlbum", endpoints.DeleteAlbum)
	}

	admins := router.Group("", middlewares.AuthMiddleware(), middlewares.AdminMiddleware())
	{
		admins.POST("/grantGlobalPermission", endpoints.GrantGlobalPermission)
		admins.POST("/revokeGlobalPermission", endpoints.RevokeGlobalPermission)
		admins.POST("/setAlbumAccess", endpoints.SetAlbumAccess)
	}

	router.Run("0.0.0.0:8080")
}
