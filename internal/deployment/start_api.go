package deployment

import (
	"data-storage-svc/internal"
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/api/endpoints"
	"data-storage-svc/internal/api/middlewares"
	"data-storage-svc/internal/api/services"
	"data-storage-svc/internal/database"
	"data-storage-svc/internal/repository"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func StartApi() {
	router := gin.Default()
	router.Use(middlewares.CORSMiddleware())

	slog.Debug("Getting mongo client")
	db := database.Mongo()

	slog.Debug("Creating repositories")
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
	albumService := services.NewAlbumService(albumRepository, mediaInAlbumRepository, albumAccessService, sharedLinkRepository, mediaRepository)
	mediaAccessService := services.NewMediaAccessService(mediaAccessRepository)
	mediaService := services.NewMediaService(mediaRepository, mediaInAlbumRepository, mediaAccessService, albumService)
	userService := services.NewUserService(userRepository)
	downloadService := services.NewDownloadService(albumRepository, downloadRepository, mediaRepository, mediaInAlbumRepository)
	sharedLinkService := services.NewSharedLinkService(sharedLinkRepository, albumAccessRepository)

	// Create middlewares
	userMiddleware := middlewares.UserMiddleware(userRepository)
	sharedLinkMiddleware := middlewares.SharedLinkMiddleware(sharedLinkRepository)

	permissionManager := common.NewPermissionsManager(albumAccessRepository, albumRepository, downloadRepository, mediaAccessRepository, mediaInAlbumRepository, mediaRepository)

	// Create endpoints
	albumEndpoint := endpoints.NewAlbumEndpoint([]gin.HandlerFunc{}, permissionManager, albumService, albumAccessService, mediaService, userService)
	mediaEndpoint := endpoints.NewMediaEndpoint([]gin.HandlerFunc{}, permissionManager, mediaService, mediaAccessService)
	userEndpoint := endpoints.NewUserEndpoint([]gin.HandlerFunc{}, permissionManager, userService)
	downloadEndpoint := endpoints.NewDownloadEndpoint([]gin.HandlerFunc{}, permissionManager, downloadService, albumAccessService)
	sharedLinkEndpoint := endpoints.NewSharedLinkEndpoint([]gin.HandlerFunc{}, permissionManager, sharedLinkService, albumService)

	endpointGroupsList := []common.EndpointGroup{
		albumEndpoint,
		mediaEndpoint,
		userEndpoint,
		downloadEndpoint,
		sharedLinkEndpoint,
	}

	api := router.Group("", userMiddleware, sharedLinkMiddleware)
	{
		for _, endpoint := range endpointGroupsList {
			endpointGroup := api.Group(endpoint.GetGroupUrl(), endpoint.GetCommonMiddlewares()...)
			for methodAndPath, finalEndpoint := range endpoint.GetEndpointsList() {
				endpointGroup.Handle(methodAndPath.Method, methodAndPath.Path, finalEndpoint...)
			}
		}
	}
	router.Run(fmt.Sprintf("%s:%d", internal.API_IP, internal.API_PORT))
}
