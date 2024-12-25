package deployment

import (
	"data-storage-svc/internal/api/endpoints"
	"data-storage-svc/internal/api/middlewares"
	"data-storage-svc/internal/api/services"
	"data-storage-svc/internal/database"
	"data-storage-svc/internal/repository"

	"github.com/gin-gonic/gin"
)

func StartApi() {
	router := gin.Default()
	router.Use(middlewares.CORSMiddleware())

	db := database.Mongo()
	// Create repositories
	albumAccessRepository := repository.NewAlbumAccessRepository(db)
	albumRepository := repository.NewAlbumRepository(db)
	mediaAccessRepository := repository.NewMediaAccessRepository(db)
	mediaInAlbumRepository := repository.NewMediaInAlbumRepository(db)
	mediaRepository := repository.NewMediaRepository(db)
	userRepository := repository.NewUserRepository(db)
	downloadRepository := repository.NewDownloadRepository(db)
	sharedLinkRepository := repository.NewSharedLinkRepository(db)

	// Create services
	albumAccessService := services.NewAlbumAccessService(albumAccessRepository)
	albumService := services.NewAlbumService(albumRepository, mediaInAlbumRepository, albumAccessService)
	mediaAccessService := services.NewMediaAccessService(mediaAccessRepository)
	mediaService := services.NewMediaService(mediaRepository, mediaInAlbumRepository, mediaAccessService, albumService)
	userService := services.NewUserService(userRepository)
	downloadService := services.NewDownloadService(albumRepository, downloadRepository, mediaRepository, mediaInAlbumRepository)
	sharedLinkService := services.NewSharedLinkService(sharedLinkRepository, albumAccessRepository)

	// Create middlewares
	sharedLinkMiddleware := middlewares.SharedLinkMiddleware(sharedLinkRepository)
	authMiddleware := middlewares.AuthMiddleware(userService)
	// Create endpoints
	albumEndpoint := endpoints.NewAlbumEndpoint(albumService, albumAccessService, mediaService, authMiddleware)
	mediaEndpoint := endpoints.NewMediaEndpoint(mediaService, mediaAccessService, authMiddleware)
	userEndpoint := endpoints.NewUserEndpoint(userService)
	downloadEndpoint := endpoints.NewDownloadEndpoint(downloadService, albumAccessService)
	sharedLinkEndpoint := endpoints.NewSharedLinkEndpoint(sharedLinkService, authMiddleware)

	// Public endpoints
	public := router.Group("")
	{
		user := public.Group("/user")
		{
			user.POST("/register", userEndpoint.Create)
			user.POST("/jwt", userEndpoint.FetchToken)
		}
	}

	// Authorized endpoints (require login)
	authorized := router.Group("", authMiddleware)
	{
		user := authorized.Group("/user")
		{
			user.POST("/logout", userEndpoint.Logout)
		}
		media := authorized.Group("/media")
		{
			media.POST("/upload", mediaEndpoint.Create)
			media.GET("/list", mediaEndpoint.List)
			media.DELETE("/delete", mediaEndpoint.Delete)
		}
		album := authorized.Group("/album")
		{
			album.POST("/create", albumEndpoint.Create)
			album.GET("/list", albumEndpoint.GetAll)
			album.DELETE("/delete", albumEndpoint.Delete)
		}
		download := authorized.Group("/download")
		{
			download.POST("/init", downloadEndpoint.InitDownload)
			download.GET("/isready", downloadEndpoint.IsReady)
			download.GET("/download", downloadEndpoint.Download)
			download.GET("/get", downloadEndpoint.Get)
		}

		sharedLink := authorized.Group("/sharedlink")
		{
			sharedLink.POST("/create", sharedLinkEndpoint.Create)
		}
	}

	authorizedOrSharedLink := router.Group("", sharedLinkMiddleware, authMiddleware)
	{
		album := authorizedOrSharedLink.Group("/album")
		{
			album.POST("/addmedia", albumEndpoint.AddMedia)
			album.GET("/get", albumEndpoint.Get)
			album.GET("/getmedias", albumEndpoint.GetMedias)
		}

		media := authorizedOrSharedLink.Group("/media")
		{
			media.GET("/get", mediaEndpoint.Get)
		}
	}

	router.Run("0.0.0.0:8080")
}
