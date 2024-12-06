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
}
type albumEndpoint struct {
	common.Endpoint
	albumService       services.AlbumService
	albumAccessService services.AlbumAccessService
	mediaService       services.MediaService
}

type CreateAlbumBody struct {
	AlbumTitle       string `json:"albumTitle"`
	AlbumDescription string `json:"albumDescription"`
}

func NewAlbumEndpoint(albumService services.AlbumService, albumAccessService services.AlbumAccessService, mediaService services.MediaService, authMiddleware gin.HandlerFunc) AlbumEndpoint {
	return albumEndpoint{Endpoint: common.NewEndpoint("album", "/album", []gin.HandlerFunc{authMiddleware}), albumService: albumService, albumAccessService: albumAccessService, mediaService: mediaService}
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

	newAlbum := &model.Album{Title: createAlbumBody.AlbumTitle, Description: createAlbumBody.AlbumDescription, CreationDate: time.Now(), AuthorId: &user.Id}

	createdId, svcErr := e.albumService.Create(newAlbum)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.IndentedJSON(http.StatusCreated, gin.H{"albumId": createdId})
}

func (e albumEndpoint) Get(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	albumId, err := utils.DecodeQueryId("albumId", c)
	if err != nil {
		return
	}

	// Check that user has authorizations to view the album
	if !e.albumAccessService.CanViewAlbum(&user.Id, albumId) {
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
	user, err := utils.GetUser(c)
	if err != nil {
		return
	}

	albumId, err := utils.DecodeQueryId("albumId", c)
	if err != nil {
		return
	}

	// Check that the user is allowed to view this album
	if !e.albumAccessService.CanViewAlbum(&user.Id, albumId) {
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
	user, err := utils.GetUser(c)
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
	if !e.albumAccessService.CanEditAlbum(&user.Id, albumId) {
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

	album, svcErr := e.albumService.GetAlbumById(albumId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	// Only the author can delete the album
	if *album.AuthorId != user.Id {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	svcErr = e.albumService.Delete(albumId)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}
	c.Status(http.StatusNoContent)
}
