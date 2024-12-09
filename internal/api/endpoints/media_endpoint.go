package endpoints

import (
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/api/services"
	"data-storage-svc/internal/utils"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MediaEndpoint interface {
	// Upload a new media
	Create(c *gin.Context)
	// List all media accessible to the given user
	List(c *gin.Context)
	// Get a specific media by id
	Get(c *gin.Context)
	// Delete a specific media by id
	Delete(c *gin.Context)
}
type mediaEndpoint struct {
	common.Endpoint
	mediaService       services.MediaService
	mediaAccessService services.MediaAccessService
}

func NewMediaEndpoint(mediaService services.MediaService, mediaAccessService services.MediaAccessService, authMiddleware gin.HandlerFunc) MediaEndpoint {
	return mediaEndpoint{Endpoint: common.NewEndpoint("album", "/album", []gin.HandlerFunc{authMiddleware}), mediaService: mediaService, mediaAccessService: mediaAccessService}
}

func (e mediaEndpoint) Create(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	contentType := c.GetHeader("Content-Type")

	if contentType != "application/octet-stream" {
		slog.Debug("Invalid content type received while upload media", "content-type", contentType)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Expected Content-Type: application/octet-stream"})
		return
	}
	fileName := c.GetHeader("Content-Disposition")
	createdId, svcErr := e.mediaService.Create(fileName, &user.Id, &c.Request.Body)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"mediaId": createdId})
}

func (e mediaEndpoint) List(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	medias, svcErr := e.mediaService.GetAllUploadedByUser(&user.Id)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.IndentedJSON(http.StatusOK, medias)
}

func (e mediaEndpoint) Get(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	// Decode the media id requested in query param
	mediaId, err := utils.DecodeQueryId("mediaId", c)
	if err != nil {
		return
	}

	// Check if user can access this media
	if !e.mediaAccessService.CanView(&user.Id, mediaId) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Get the media meta-info
	media, svcErr := e.mediaService.GetById(mediaId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}
	// Get the media data
	mimeType, data, svcErr := e.mediaService.GetData(*media.StorageFileName)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.Data(http.StatusOK, *mimeType, data)
}

func (e mediaEndpoint) Delete(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	// Decode the media id requested in query param
	mediaId, err := utils.DecodeQueryId("mediaId", c)
	if err != nil {
		return
	}

	// Get the media meta-info
	media, svcErr := e.mediaService.GetById(mediaId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	// Only the owner can delete a media
	if *media.UploadedBy != user.Id {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Actually delete the media
	svcErr = e.mediaService.Delete(&media.Id)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.Status(http.StatusNoContent)
}
