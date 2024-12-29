package endpoints

import (
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/api/services"
	"data-storage-svc/internal/model"
	"data-storage-svc/internal/utils"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AlbumEndpoint interface {
	// Create an album
	Create(c *gin.Context)
	// Get a specific album, by id
	Get(c *gin.Context)
	// Get all albums accessible for the user
	GetAll(c *gin.Context)
	// Get all medias in the given album
	GetMedias(c *gin.Context)
	// Add a media in the album
	AddMedia(c *gin.Context)
	// Delete an album (not the underlying medias)
	Delete(c *gin.Context)
	// Return the list of users who can access this album
	SharedWith(c *gin.Context)
	// Modify access to the album for the given user
	Access(c *gin.Context)
}
type albumEndpoint struct {
	common.Endpoint
	albumService       services.AlbumService
	albumAccessService services.AlbumAccessService
	mediaService       services.MediaService
	userService        services.UserService
}

type CreateAlbumBody struct {
	AlbumTitle       string `json:"albumTitle"`
	AlbumDescription string `json:"albumDescription"`
}

func NewAlbumEndpoint(permissionsManager common.PermissionsManager, albumService services.AlbumService, albumAccessService services.AlbumAccessService, mediaService services.MediaService, authMiddleware gin.HandlerFunc, userService services.UserService) AlbumEndpoint {
	return albumEndpoint{Endpoint: common.NewEndpoint("album", "/album", []gin.HandlerFunc{authMiddleware}, permissionsManager), albumService: albumService, albumAccessService: albumAccessService, mediaService: mediaService, userService: userService}
}

func (e albumEndpoint) Create(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}
	var createAlbumBody CreateAlbumBody
	if err := c.BindJSON(&createAlbumBody); err != nil {
		slog.Debug("Couldn't decode create album body", "error", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if !e.GetPermissionsManager().CanCreateAlbum(user) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	newAlbum := &model.Album{Title: createAlbumBody.AlbumTitle, Description: createAlbumBody.AlbumDescription, CreationDate: time.Now(), AuthorId: &user.Id}

	createdId, svcErr := e.albumService.Create(newAlbum)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.IndentedJSON(http.StatusCreated, gin.H{"albumId": createdId})
}

func (e albumEndpoint) Get(c *gin.Context) {
	user, sharedLink, err := utils.GetUserOrSharedLink(c)
	if err != nil {
		return
	}

	albumId, err := utils.DecodeQueryId("albumId", c)
	if err != nil {
		return
	}

	// Check that user has authorizations to view the album
	if !e.GetPermissionsManager().CanGetAlbum(user, albumId, sharedLink) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Fetch it
	album, svcErr := e.albumService.GetAlbumById(albumId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.IndentedJSON(http.StatusOK, album)
}

func (e albumEndpoint) GetAll(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	albums, svcErr := e.albumService.GetAllAlbumsForUser(&user.Id)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.IndentedJSON(http.StatusOK, albums)
}

func (e albumEndpoint) GetMedias(c *gin.Context) {
	user, sharedLink, err := utils.GetUserOrSharedLink(c)
	if err != nil {
		return
	}

	albumId, err := utils.DecodeQueryId("albumId", c)
	if err != nil {
		return
	}

	// Check that the user is allowed to view this album
	if !e.GetPermissionsManager().CanGetAllMediasForAlbum(user, albumId, sharedLink) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// List all medias in the album
	mediasInAlbum, svcErr := e.albumService.GetMedias(albumId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	type ResponseItem struct {
		Media *model.Media        `json:"media"`
		Link  *model.MediaInAlbum `json:"link"`
	}

	// Fetch media content and merge with access infos
	var result []ResponseItem = make([]ResponseItem, 0)
	for _, m := range mediasInAlbum {
		respItem := ResponseItem{Link: &m}
		media, svcErr := e.mediaService.GetById(m.MediaId)
		if svcErr != nil {
			svcErr.Apply(c)
			return
		}
		respItem.Media = media
		result = append(result, respItem)
	}

	c.IndentedJSON(http.StatusOK, result)
}

type AddMediaToAlbumBody struct {
	MediaId string `json:"mediaId"`
	AlbumId string `json:"albumId"`
}

func (e albumEndpoint) AddMedia(c *gin.Context) {
	user, sharedLink, err := utils.GetUserOrSharedLink(c)
	if err != nil {
		return
	}

	var addMediaToAlbumBody AddMediaToAlbumBody
	if err := c.BindJSON(&addMediaToAlbumBody); err != nil {
		slog.Debug("Couldn't decode body", "error", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Decoding media ID from the request body
	mediaId, svcErr := utils.DecodeBodyId(addMediaToAlbumBody.MediaId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	// Get the media (check if it exists)
	_, svcErr = e.mediaService.GetById(mediaId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	// Decoding album ID from the request body
	albumId, svcErr := utils.DecodeBodyId(addMediaToAlbumBody.AlbumId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	// Check if user is allowed to edit the album
	if !e.GetPermissionsManager().CanAddMediaToAlbum(user, albumId, sharedLink) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Add media to album
	addTime := time.Now()
	mediaInAlbum := &model.MediaInAlbum{MediaId: mediaId, AlbumId: albumId, AddedBy: &user.Id, AddedDate: &addTime}
	svcErr = e.albumService.AddMedia(mediaInAlbum)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.Status(http.StatusOK)
}

func (e albumEndpoint) Delete(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}
	albumId, err := utils.DecodeQueryId("albumId", c)
	if err != nil {
		return
	}

	// Only the author can delete the album
	if !e.GetPermissionsManager().CanDeleteAlbum(user, albumId) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	svcErr := e.albumService.Delete(albumId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}
	c.Status(http.StatusNoContent)
}

func (e albumEndpoint) SharedWith(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}
	albumId, err := utils.DecodeQueryId("albumId", c)
	if err != nil {
		return
	}

	if !e.GetPermissionsManager().CanListAlbumAccesses(user, albumId) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	accesses, svcErr := e.albumAccessService.GetAllAccesses(albumId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	type Result struct {
		Email   string `json:"email"`
		CanEdit bool   `json:"canEdit"`
	}

	var result []Result = make([]Result, 0)

	for _, access := range accesses {
		userShared, err := e.userService.GetById(*access.UserId)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// Do not include the user's access as it is implicit
		if userShared.Id.Hex() != user.Id.Hex() {
			result = append(result, Result{Email: userShared.Email, CanEdit: access.CanEdit})
		}
	}

	c.IndentedJSON(http.StatusOK, result)
}

type AccessBody struct {
	AlbumId   string `json:"albumId"`
	UserEmail string `json:"email"`
	AllowEdit bool   `json:"allowEdit,omitempty"`
}

func (e albumEndpoint) Access(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	var accessBody AccessBody
	if err := c.BindJSON(&accessBody); err != nil {
		slog.Debug("Couldn't decode body", "error", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	albumId, svcErr := utils.DecodeBodyId(accessBody.AlbumId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	if !e.GetPermissionsManager().CanEditAlbumAccesses(user, albumId) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Get user to share/unshare the album with
	userToShareWith, svcErr := e.userService.GetByEmail(accessBody.UserEmail)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	// Cannot change the owner's access
	if userToShareWith.Id.Hex() == user.Id.Hex() {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if c.Request.Method == "POST" {
		// Create/modify the access
		svcErr = e.albumAccessService.GrantAccess(&userToShareWith.Id, albumId, accessBody.AllowEdit)
		if svcErr != nil {
			svcErr.Apply(c)
			return
		}
	} else if c.Request.Method == "DELETE" {
		// Revoke an existing access
		svcErr = e.albumAccessService.RevokeAccess(&userToShareWith.Id, albumId)
		if svcErr != nil {
			svcErr.Apply(c)
			return
		}
	} else {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	c.Status(http.StatusOK)
}
