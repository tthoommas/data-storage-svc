package endpoints

import (
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/api/middlewares"
	"data-storage-svc/internal/api/services"
	"data-storage-svc/internal/utils"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DownloadEndpoint interface {
	common.EndpointGroup
	// Init a download (ie. asynchronous zip file creation)
	InitDownload(c *gin.Context)
	// Download a previously created zip file
	Download(c *gin.Context)
	// Get a specifc download meta-data
	Get(c *gin.Context)
}
type downloadEndpoint struct {
	common.EndpointGroup
	downloadService    services.DownloadService
	albumAccessService services.AlbumAccessService
}

func NewDownloadEndpoint(
	// Common dependencies
	commonMiddlewares []gin.HandlerFunc,
	permissionsManager common.PermissionsManager,
	// Service dependencies
	downloadService services.DownloadService,
	albumAccessService services.AlbumAccessService,
) DownloadEndpoint {
	downloadEndpoint := downloadEndpoint{
		downloadService:    downloadService,
		albumAccessService: albumAccessService,
	}

	endpoint := common.NewEndpoint(
		"Downloads",
		"/download",
		commonMiddlewares,
		map[common.MethodPath][]gin.HandlerFunc{
			// Common album edition actions
			{Method: "POST", Path: ""}:                 {downloadEndpoint.InitDownload},
			{Method: "GET", Path: "/:downloadId/meta"}: {middlewares.PathParamIdMiddleware("downloadId"), downloadEndpoint.Get},
			{Method: "GET", Path: "/:downloadId/data"}: {middlewares.PathParamIdMiddleware("downloadId"), downloadEndpoint.Download},
		},
		permissionsManager,
	)

	downloadEndpoint.EndpointGroup = endpoint
	return &downloadEndpoint
}

type DownloadAlbumBody struct {
	AlbumId    string   `json:"albumId"`
	Everything bool     `json:"everything"`
	MediaList  []string `json:"mediaList"`
	// TODO allow to specify media quality MediasQuality model.MediaQuality `json:"mediasQuality"`
}

func (e *downloadEndpoint) InitDownload(c *gin.Context) {
	user, sharedLink, err := utils.GetUserOrSharedLink(c)
	if err != nil {
		return
	}

	var downloadAlbumBody DownloadAlbumBody
	if err := c.BindJSON(&downloadAlbumBody); err != nil {
		slog.Debug("Couldn't decode body", "error", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Decoding album ID from the request body
	albumId, svcErr := utils.DecodeBodyId(downloadAlbumBody.AlbumId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	if !e.GetPermissionsManager().CanInitDownloadForAlbum(user, albumId, sharedLink) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if downloadAlbumBody.Everything {
		// Download the full album (i.e. all medias inside the album)
		downloadId, err := e.downloadService.InitDownload(albumId, &user.Id)
		if err != nil {
			err.Apply(c)
			return
		}
		// Return the created download ID
		c.IndentedJSON(http.StatusCreated, gin.H{"downloadId": downloadId.Hex()})
	} else {
		c.AbortWithStatus(http.StatusNotImplemented)
	}
}

func (e *downloadEndpoint) Download(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	// Decode the download id requested in query param
	downloadId := utils.GetIdFromContext("downloadId", c)

	download, svcErr := e.downloadService.Get(&downloadId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	if !e.GetPermissionsManager().CanConsumeDownload(user, &downloadId) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	data, svcErr := e.downloadService.GetData(&downloadId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", download.DownloadName))
	c.Data(http.StatusOK, "application/x-zip", data)
}

func (e *downloadEndpoint) Get(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	// Decode the download id requested in query param
	downloadId := utils.GetIdFromContext("downloadId", c)

	if !e.GetPermissionsManager().CanGetDownload(user, &downloadId) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	download, svcErr := e.downloadService.Get(&downloadId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.IndentedJSON(http.StatusOK, download)
}
