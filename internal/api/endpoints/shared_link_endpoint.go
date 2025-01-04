package endpoints

import (
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/api/services"
	"data-storage-svc/internal/utils"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type SharedLinkEndpoint interface {
	common.EndpointGroup
	// Create a new shared link
	Create(c *gin.Context)
	// List all shared links for a given album
	List(c *gin.Context)
	// Delete a given shared link
	Delete(c *gin.Context)
	// Update a given shared link
	Update(c *gin.Context)
}
type sharedLinkEndpoint struct {
	common.EndpointGroup
	sharedLinkService services.SharedLinkService
	albumService      services.AlbumService
}

func NewSharedLinkEndpoint(
	// Common dependencies
	commonMiddlewares []gin.HandlerFunc,
	permissionsManager common.PermissionsManager,
	//Services dependencies
	sharedLinkService services.SharedLinkService,
	albumService services.AlbumService,
) SharedLinkEndpoint {
	sharedLinkEndpoint := sharedLinkEndpoint{
		sharedLinkService: sharedLinkService,
		albumService:      albumService,
	}

	endpoint := common.NewEndpoint(
		"Shared links",
		"/sharedlink",
		commonMiddlewares,
		map[common.MethodPath][]gin.HandlerFunc{
			// Common album edition actions
			{Method: "POST", Path: "/"}:                {sharedLinkEndpoint.Create},
			{Method: "GET", Path: "/"}:                 {sharedLinkEndpoint.List},
			{Method: "DELETE", Path: "/:sharedLinkId"}: {sharedLinkEndpoint.Delete},
			{Method: "PATCH", Path: "/:sharedLinkId"}:  {sharedLinkEndpoint.Update},
		},
		permissionsManager,
	)

	sharedLinkEndpoint.EndpointGroup = endpoint
	return sharedLinkEndpoint
}

type CreateBody struct {
	AlbumId   string `json:"albumId"`
	TTL       int    `json:"ttl"`
	AllowEdit bool   `json:"allowEdit"`
}

func (e sharedLinkEndpoint) Create(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	var createLinkBody CreateBody
	if err := c.BindJSON(&createLinkBody); err != nil {
		slog.Debug("Couldn't decode body", "error", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	albumId, svcErr := utils.DecodeBodyId(createLinkBody.AlbumId)
	if svcErr != nil || albumId == nil {
		svcErr.Apply(c)
		return
	}

	if !e.GetPermissionsManager().CanCreateSharedLink(user, albumId) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(time.Second * time.Duration(createLinkBody.TTL))
	link, svcErr := e.sharedLinkService.Create(*albumId, user.Id, expirationTime, createLinkBody.AllowEdit)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.IndentedJSON(http.StatusCreated, gin.H{"sharedLinkToken": link.Token})
}

func (e sharedLinkEndpoint) List(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	albumId, err := utils.DecodeQueryId("albumId", c)
	if err != nil || albumId == nil {
		return
	}

	if !e.GetPermissionsManager().CanListSharedLinks(user, albumId) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	sharedLinks, svcErr := e.sharedLinkService.List(*albumId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.IndentedJSON(http.StatusOK, sharedLinks)
}

func (e sharedLinkEndpoint) Delete(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	sharedLink, svcErr := e.sharedLinkService.GetByToken(c.Query("token"))
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	// Only the creator can delete its link
	if !e.GetPermissionsManager().CanDeleteSharedLink(user, sharedLink) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	svcErr = e.sharedLinkService.Delete(sharedLink.Id)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.Status(http.StatusNoContent)
}

type UpdateBody struct {
	Token     string `json:"token"`
	AllowEdit bool   `json:"allowEdit"`
}

func (e sharedLinkEndpoint) Update(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	var updateBody UpdateBody
	if err := c.BindJSON(&updateBody); err != nil {
		slog.Debug("Couldn't decode body", "error", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	sharedLink, svcErr := e.sharedLinkService.GetByToken(updateBody.Token)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	// Only creator can update its link
	if !e.GetPermissionsManager().CanUpdateSharedLink(user, sharedLink) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	svcErr = e.sharedLinkService.Update(sharedLink.Id, updateBody.AllowEdit)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.Status(http.StatusOK)
}
