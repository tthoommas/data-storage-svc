package endpoints

import (
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/api/services"
	"data-storage-svc/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SharedLinkEndpoint interface {
	// Create a new shared link
	Create(c *gin.Context)
}
type sharedLinkEndpoint struct {
	common.Endpoint
	sharedLinkService services.SharedLinkService
}

func NewSharedLinkEndpoint(sharedLinkService services.SharedLinkService, authMiddleware gin.HandlerFunc) SharedLinkEndpoint {
	return sharedLinkEndpoint{Endpoint: common.NewEndpoint("sharedlink", "/sharedlink", []gin.HandlerFunc{authMiddleware}), sharedLinkService: sharedLinkService}
}

func (e sharedLinkEndpoint) Create(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	albumId, err := utils.DecodeQueryId("albumId", c)
	if err != nil || albumId == nil {
		return
	}

	link, svcErr := e.sharedLinkService.Create(*albumId, user.Id)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.IndentedJSON(http.StatusCreated, gin.H{"sharedLinkToken": link.Token})
}
