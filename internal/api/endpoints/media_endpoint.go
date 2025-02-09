package endpoints

import (
	"bytes"
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/api/middlewares"
	"data-storage-svc/internal/api/services"
	"data-storage-svc/internal/utils"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MediaEndpoint interface {
	common.EndpointGroup
	// Upload a new media
	Create(c *gin.Context)
	// List all media accessible to the given user
	List(c *gin.Context)
	// Get a specific media by id
	Get(c *gin.Context)
	// Get media meta data by id
	GetMetaData(c *gin.Context)
	// Delete a specific media by id
	Delete(c *gin.Context)
}
type mediaEndpoint struct {
	common.EndpointGroup
	mediaService       services.MediaService
	mediaAccessService services.MediaAccessService
}

func NewMediaEndpoint(
	commonMiddlewares []gin.HandlerFunc,
	permissionsManager common.PermissionsManager,
	mediaService services.MediaService,
	mediaAccessService services.MediaAccessService,
) MediaEndpoint {
	mediaEndpoint := mediaEndpoint{
		mediaService:       mediaService,
		mediaAccessService: mediaAccessService,
	}

	endpoint := common.NewEndpoint(
		"Medias",
		"/media",
		commonMiddlewares,
		map[common.MethodPath][]gin.HandlerFunc{
			{Method: "POST", Path: ""}:              {mediaEndpoint.Create},
			{Method: "GET", Path: ""}:               {mediaEndpoint.List},
			{Method: "GET", Path: "/:mediaId"}:      {middlewares.PathParamIdMiddleware("mediaId"), mediaEndpoint.Get},
			{Method: "GET", Path: "/:mediaId/meta"}: {middlewares.PathParamIdMiddleware("mediaId"), mediaEndpoint.GetMetaData},
			{Method: "DELETE", Path: "/:mediaId"}:   {middlewares.PathParamIdMiddleware("mediaId"), mediaEndpoint.Delete},
		},
		permissionsManager,
	)

	mediaEndpoint.EndpointGroup = endpoint
	return &mediaEndpoint
}

func (e *mediaEndpoint) Create(c *gin.Context) {
	user, sharedLink, err := utils.GetUserOrSharedLink(c)
	if err != nil {
		return
	}

	if !e.GetPermissionsManager().CanCreateMedia(user, sharedLink) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	contentType := c.GetHeader("Content-Type")

	if contentType != "application/octet-stream" {
		slog.Debug("Invalid content type received while upload media", "content-type", contentType)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Expected Content-Type: application/octet-stream"})
		return
	}
	fileName := c.GetHeader("Content-Disposition")

	fileName, err = utils.ToUTF8(fileName)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		slog.Debug("couldn't encode filename as utf-8", "err", err)
		return
	}
	addedById, isSharedLink, err := utils.GetUserIdOrLinkId(user, sharedLink)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	createdId, svcErr := e.mediaService.Create(fileName, addedById, isSharedLink, &c.Request.Body)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"mediaId": createdId})
}

func isValidUTF8(s string) bool {
	return bytes.Equal([]byte(s), []byte(string([]rune(s))))
}

func (e *mediaEndpoint) List(c *gin.Context) {
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

func (e *mediaEndpoint) Get(c *gin.Context) {
	user, sharedLink, err := utils.GetUserOrSharedLink(c)
	if err != nil {
		return
	}
	mediaId := utils.GetIdFromContext("mediaId", c)

	// Decode the requested size in query param (if any)
	compressedQualityRaw := c.DefaultQuery("compressed", "true")
	compressedQuality := true
	if compressedQualityRaw == "false" {
		compressedQuality = false
	}

	if !e.GetPermissionsManager().CanGetMedia(user, &mediaId, sharedLink) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Get the media meta-info
	media, svcErr := e.mediaService.GetById(&mediaId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}
	// Get the media data
	mimeType, data, svcErr := e.mediaService.GetData(*media.StorageFileName, compressedQuality)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.Data(http.StatusOK, *mimeType, data)
}

func (e *mediaEndpoint) GetMetaData(c *gin.Context) {
	user, sharedLink, err := utils.GetUserOrSharedLink(c)
	if err != nil {
		return
	}
	mediaId := utils.GetIdFromContext("mediaId", c)

	if !e.GetPermissionsManager().CanGetMedia(user, &mediaId, sharedLink) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	metaDatas, svcErr := e.mediaService.GetMetaData(&mediaId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}
	c.IndentedJSON(http.StatusOK, metaDatas)
}

func (e *mediaEndpoint) Delete(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	mediaId := utils.GetIdFromContext("mediaId", c)

	// Get the media meta-info
	media, svcErr := e.mediaService.GetById(&mediaId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	if !e.GetPermissionsManager().CanDeleteMedia(user, &mediaId) {
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
