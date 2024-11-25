package deployment

import (
	"data-storage-svc/internal/api/endpoints"
	"data-storage-svc/internal/api/middlewares"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func StartApi() {
	slog.Info("Starting data-storage service...")
	startMongoDB()
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
	}

	admins := router.Group("", middlewares.AuthMiddleware(), middlewares.AdminMiddleware())
	{
		admins.POST("/grantGlobalPermission", endpoints.GrantGlobalPermission)
		admins.POST("/revokeGlobalPermission", endpoints.RevokeGlobalPermission)
		admins.POST("/setAlbumAccess", endpoints.SetAlbumAccess)
	}

	router.Run("0.0.0.0:8080")
}
