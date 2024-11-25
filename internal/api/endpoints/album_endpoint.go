package endpoints

import (
	"data-storage-svc/internal/database"
	"data-storage-svc/internal/model"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetAlbumById(c *gin.Context) {
	rawAlbumId := c.Query("albumId")
	albumId, err := primitive.ObjectIDFromHex(rawAlbumId)
	if rawAlbumId == "" || err != nil {
		slog.Debug("Fetching album with invalid ID", "rawAlbumId", rawAlbumId)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid album ID"})
		return
	}

	rawUser, _ := c.Get("user")
	user := rawUser.(*model.User)

	if !database.CanUserAccessAlbum(user.Id, &albumId) {
		slog.Debug("User is not allowed to access this album", "user", *user.Email, "album", albumId)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	album, err := database.GetAlbumById(&albumId)
	if err != nil {
		slog.Debug("Couldn't find album.", "albumId", albumId, "error", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.IndentedJSON(http.StatusOK, album)
}

func GetMyAlbums(c *gin.Context) {
	rawUser, _ := c.Get("user")
	user := rawUser.(*model.User)
	albumsAccesses, err := database.GetAllAlbumsForUser(user.Id)
	if err != nil {
		slog.Debug("Couldn't get albums accesses for user.", "user", *user.Email, "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	albums := make([]model.Album, 0)
	for _, albumAccess := range albumsAccesses {
		album, err := database.GetAlbumById(albumAccess.AlbumId)
		if err != nil {
			slog.Debug("Couldn't find album", "albumId", albumAccess.AlbumId)
		} else {
			albums = append(albums, *album)
		}
	}
	c.IndentedJSON(http.StatusOK, albums)
}

func GetMediasInAlbum(c *gin.Context) {
	rawAlbumId := c.Query("albumId")
	albumId, err := primitive.ObjectIDFromHex(rawAlbumId)
	if rawAlbumId == "" || err != nil {
		slog.Debug("Fetching album with invalid ID", "rawAlbumId", rawAlbumId)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid album ID"})
		return
	}

	rawUser, _ := c.Get("user")
	user := rawUser.(*model.User)

	if !database.CanUserAccessAlbum(user.Id, &albumId) {
		slog.Debug("User is not allowed to access this album", "user", *user.Email, "album", albumId)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	mediasInAlbum, err := database.GetAllMediasInAlbum(&albumId)
	if err != nil {
		slog.Debug("Couldn't get all medias in album.", "albumId", albumId, "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	medias := make([]model.Media, 0)
	for _, mediaInAlbum := range mediasInAlbum {
		media, err := database.GetMedia(mediaInAlbum.MediaId)
		if err != nil {
			slog.Debug("Coulnd't find media in album", "eror", err, "mediaId", *mediaInAlbum.MediaId)
		} else {
			medias = append(medias, *media)
		}
	}

	c.IndentedJSON(http.StatusOK, medias)
}

type CreateAlbumBody struct {
	AlbumTitle       string `json:"albumTitle"`
	AlbumDescription string `json:"albumDescription"`
}

func CreateAlbum(c *gin.Context) {
	var createAlbumBody CreateAlbumBody
	if err := c.BindJSON(&createAlbumBody); err != nil {
		slog.Debug("Couldn't decode create album body", "error", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if len(createAlbumBody.AlbumTitle) == 0 {
		slog.Debug("Cannot create album with empty title")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	user, _ := c.Get("user")
	uploadedBy := user.(*model.User)
	newAlbum := model.Album{Title: createAlbumBody.AlbumTitle, Description: createAlbumBody.AlbumDescription, CreationDate: time.Now(), AuthorId: uploadedBy.Id}
	insertedID, err := database.CreateAlbum(&newAlbum)
	if err != nil {
		slog.Debug("Couldn't create album", "title", newAlbum.Title, "author", uploadedBy.Email, "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if err := database.GrantAccessToAlbum(uploadedBy.Id, insertedID, true); err != nil {
		slog.Error("Couldn't grant access to album author", "createdAlbumId", insertedID, "user", *uploadedBy.Email, "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.IndentedJSON(http.StatusCreated, gin.H{"albumId": insertedID})
}

type AddMediaToAlbumBody struct {
	MediaId string `json:"mediaId"`
	AlbumId string `json:"albumId"`
}

func AddMediaToAlbum(c *gin.Context) {
	var addMediaToAlbumBody AddMediaToAlbumBody
	if err := c.BindJSON(&addMediaToAlbumBody); err != nil {
		slog.Debug("Couldn't decode body", "error", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Getting user
	rawUser, _ := c.Get("user")
	user := rawUser.(*model.User)

	// Decoding media ID from the request body
	mediaId, err := primitive.ObjectIDFromHex(addMediaToAlbumBody.MediaId)
	if err != nil {
		slog.Debug("Add media to album with invalid media ID", "user", *user.Email, "rawMediaId", addMediaToAlbumBody.MediaId)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	// Check that this media exists
	if _, err := database.GetMedia(&mediaId); err != nil {
		slog.Debug("Media do not exists. Cannot add to album", "mediaId", mediaId)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// Decoding album ID from the request body
	albumId, err := primitive.ObjectIDFromHex(addMediaToAlbumBody.AlbumId)
	if err != nil {
		slog.Debug("Adding media to album with invalid album ID", "user", *user.Email, "rawAlbumId", addMediaToAlbumBody.AlbumId)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid album ID"})
		return
	}

	// Check if user is allowed to edit the album
	if !database.CanUserEditAlbum(user.Id, &albumId) {
		slog.Debug("User is not allowed to edit this album.", "user", *user.Email, "albumId", albumId)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Add media to album
	addTime := time.Now()
	addErr := database.AddMediaToAlbum(&model.MediaInAlbum{MediaId: &mediaId, AlbumId: &albumId, AddedBy: user.Id, AddedDate: &addTime})
	if addErr != nil {
		slog.Debug("Couldn't add media to album", "mediaId", mediaId, "albumId", albumId, "user", *user.Email, "error", addErr)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func RemoveMediaFromAlbum(c *gin.Context) {

}

func DeleteAlbum(c *gin.Context) {
	rawAlbumId := c.Query("albumId")
	albumId, err := primitive.ObjectIDFromHex(rawAlbumId)
	if rawAlbumId == "" || err != nil {
		slog.Debug("Fetching album with invalid ID", "rawAlbumId", rawAlbumId)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid album ID"})
		return
	}

	rawUser, _ := c.Get("user")
	user := rawUser.(*model.User)

	if !database.CanUserDeleteAlbum(user.Id, &albumId) {
		slog.Debug("User is not allowed to delete this album", "user", *user.Email, "album", albumId)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

}

type SetAlbumAccessBody struct {
	AlbumId      string `json:"albumId"`
	TargetUserId string `json:"targetUserId"`
	CanView      bool   `json:"canView"`
	CanEdit      bool   `json:"canEdit"`
}

func SetAlbumAccess(c *gin.Context) {
	var setAlbumAccessBody SetAlbumAccessBody
	if err := c.BindJSON(&setAlbumAccessBody); err != nil {
		slog.Debug("Couldn't decode SetAlbumAccessBody", "error", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Getting user
	rawUser, _ := c.Get("user")
	user := rawUser.(*model.User)

	// Decoding album ID from the request body
	albumId, err := primitive.ObjectIDFromHex(setAlbumAccessBody.AlbumId)
	if err != nil {
		slog.Debug("Set album access invalid albumId", "user", *user.Email, "rawAlbumId", setAlbumAccessBody.AlbumId)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid album ID"})
		return
	}

	// Check that this album exists
	album, err := database.GetAlbumById(&albumId)
	if album == nil || err != nil {
		slog.Debug("Album not found. Cannot set access.", "error", err, "albumId", albumId)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// Decoding targetUser ID from the request body
	targetUserId, err := primitive.ObjectIDFromHex(setAlbumAccessBody.TargetUserId)
	if err != nil {
		slog.Debug("Set album access invalid targetUserId", "user", *user.Email, "rawTargetUserId", setAlbumAccessBody.TargetUserId)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid targetUser ID"})
		return
	}

	// Check that the targetUser exists
	targetUser, err := database.FindUserById(&targetUserId)
	if targetUser == nil || err != nil {
		slog.Debug("Target user not found. Cannot set access.", "error", err, "userId", targetUserId)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if !setAlbumAccessBody.CanView && !setAlbumAccessBody.CanEdit {
		// Revoking all permissions for targetUser on this album
		if err := database.RevokeAccessToAlbum(&targetUserId, &albumId); err != nil {
			slog.Debug("Couldn't revoke access to album", "targetUserId", targetUserId, "albumId", albumId, "error", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
		return
	}

	if err := database.GrantAccessToAlbum(&targetUserId, &albumId, setAlbumAccessBody.CanEdit); err != nil {
		slog.Debug("Couldn't grant access to album", "targetUserId", targetUserId, "albumId", albumId, "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
