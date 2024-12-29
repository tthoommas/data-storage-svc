package endpoints

import (
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PermissionsEndpoint interface {
	CanDeleteAlbum(c *gin.Context)
}

type permissionsEndpoint struct {
	common.Endpoint
}

func NewPermissionsEndpoint(permissionsManager common.PermissionsManager) PermissionsEndpoint {
	return permissionsEndpoint{Endpoint: common.NewEndpoint("album", "/album", []gin.HandlerFunc{}, permissionsManager)}
}

type Handler struct {
	handleFunction func()
}

func (e permissionsEndpoint) CanDeleteAlbum(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}
	albumId, err := utils.DecodeQueryId("albumId", c)
	if err != nil {
		return
	}
	if e.GetPermissionsManager().CanDeleteAlbum(user, albumId) {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusUnauthorized)
	}

}
