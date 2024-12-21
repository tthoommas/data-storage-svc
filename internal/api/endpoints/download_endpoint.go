package endpoints

import (
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/api/services"
	"data-storage-svc/internal/utils"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DownloadEndpoint interface {
	// Init a download (ie. asynchronous zip file creation)
	InitDownload(c *gin.Context)
	// Check if a zip file is ready to be downloaded
	IsReady(c *gin.Context)
	// Download a previously created zip file
	Download(c *gin.Context)
	// Get a specifc download meta-data
	Get(c *gin.Context)
}
type downloadEndpoint struct {
	common.Endpoint
	downloadService    services.DownloadService
	albumAccessService services.AlbumAccessService
}

func NewDownloadEndpoint(downloadService services.DownloadService, albumAccessService services.AlbumAccessService) DownloadEndpoint {
	return downloadEndpoint{Endpoint: common.NewEndpoint("download", "/download", []gin.HandlerFunc{}), downloadService: downloadService, albumAccessService: albumAccessService}
}

type DownloadAlbumBody struct {
	AlbumId    string   `json:"albumId"`
	Everything bool     `json:"everything"`
	MediaList  []string `json:"mediaList"`
	// TODO allow to specify media quality MediasQuality model.MediaQuality `json:"mediasQuality"`
}

func (e downloadEndpoint) InitDownload(c *gin.Context) {
	user, err := utils.GetUser(c)
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

	// Check that the user can access this album
	if !e.albumAccessService.CanViewAlbum(&user.Id, albumId) {
		c.Status(http.StatusUnauthorized)
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

func (e downloadEndpoint) IsReady(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	// Decode the download id requested in query param
	downloadId, err := utils.DecodeQueryId("downloadId", c)
	if err != nil {
		return
	}

	download, svcErr := e.downloadService.Get(downloadId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	// Check that this user has created this download (downloads are private)
	if download.Initiator.Hex() != user.Id.Hex() {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"isReady": download.IsReady})
}

func (e downloadEndpoint) Download(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	// Decode the download id requested in query param
	downloadId, err := utils.DecodeQueryId("downloadId", c)
	if err != nil {
		return
	}

	download, svcErr := e.downloadService.Get(downloadId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}
	// Check that this user has created this download (downloads are private)
	if download.Initiator.Hex() != user.Id.Hex() {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	data, svcErr := e.downloadService.GetData(downloadId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", download.DownloadName))
	c.Data(http.StatusOK, "application/x-zip", data)
}

func (e downloadEndpoint) Get(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	// Decode the download id requested in query param
	downloadId, err := utils.DecodeQueryId("downloadId", c)
	if err != nil {
		return
	}

	download, svcErr := e.downloadService.Get(downloadId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	// Check that user can access this download
	if download.Initiator.Hex() != user.Id.Hex() {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.IndentedJSON(http.StatusOK, download)
}
