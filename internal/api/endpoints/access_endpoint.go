package endpoints

import (
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/api/services"
	"data-storage-svc/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AccessEndpoint interface {
	// Check if the user can view a given album
	CanViewAlbum(c *gin.Context)
	// Check if the user can edit a given album
	CanEditAlbum(c *gin.Context)
	// Check if the user can view a given media
	CanViewMedia(c *gin.Context)
}
type accessEndpoint struct {
	common.Endpoint
	albumAccessService services.AlbumAccessService
	mediaAccessService services.MediaAccessService
}

func NewAccessEndpoint(albumAccessService services.AlbumAccessService, mediaAccessService services.MediaAccessService) AccessEndpoint {
	return accessEndpoint{Endpoint: common.NewEndpoint("access", "/access", []gin.HandlerFunc{}), albumAccessService: albumAccessService, mediaAccessService: mediaAccessService}
}

func (e accessEndpoint) CanViewAlbum(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	albumId, err := utils.DecodeQueryId("albumId", c)
	if err != nil {
		return
	}

	if e.albumAccessService.CanViewAlbum(&user.Id, albumId) {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusUnauthorized)
	}
}

func (e accessEndpoint) CanEditAlbum(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	albumId, err := utils.DecodeQueryId("albumId", c)
	if err != nil {
		return
	}

	if e.albumAccessService.CanEditAlbum(&user.Id, albumId) {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusUnauthorized)
	}
}

func (e accessEndpoint) CanViewMedia(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	mediaId, err := utils.DecodeQueryId("mediaId", c)
	if err != nil {
		return
	}

	if e.mediaAccessService.CanView(&user.Id, mediaId) {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusUnauthorized)
	}
}
