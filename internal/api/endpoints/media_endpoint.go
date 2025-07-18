package endpoints

import (
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/api/middlewares"
	"data-storage-svc/internal/api/services"
	"data-storage-svc/internal/utils"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tus/tusd/v2/pkg/filelocker"
	"github.com/tus/tusd/v2/pkg/filestore"
	"github.com/tus/tusd/v2/pkg/handler"
	tusd "github.com/tus/tusd/v2/pkg/handler"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MediaEndpoint interface {
	common.EndpointGroup
	// Hook called before file upload starts
	PreCreate(hook tusd.HookEvent) (tusd.HTTPResponse, tusd.FileInfoChanges, error)
	// Hook called before final http response is sent
	PreFinish(hook tusd.HookEvent) (tusd.HTTPResponse, error)
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

	// Create the TUS store and locker
	targetFolderOriginal, err := utils.GetDataDir("originalMedias")

	store := filestore.New(targetFolderOriginal)
	locker := filelocker.New(targetFolderOriginal)
	composer := tusd.NewStoreComposer()
	store.UseIn(composer)
	locker.UseIn(composer)

	// Create an unrouted handler, we will route manually with gin
	handler, err := tusd.NewUnroutedHandler(tusd.Config{
		BasePath:                "/api/media/chunkupload",
		StoreComposer:           composer,
		NotifyCompleteUploads:   false,
		RespectForwardedHeaders: true,
		PreUploadCreateCallback: func(hook tusd.HookEvent) (tusd.HTTPResponse, tusd.FileInfoChanges, error) {
			return mediaEndpoint.PreCreate(hook)
		},
		PreFinishResponseCallback: func(hook tusd.HookEvent) (tusd.HTTPResponse, error) {
			return mediaEndpoint.PreFinish(hook)
		},
	})

	if err != nil {
		log.Fatalf("unable to create handler: %s", err)
	}

	endpoint := common.NewEndpoint(
		"Medias",
		"/media",
		commonMiddlewares,
		map[common.MethodPath][]gin.HandlerFunc{
			{Method: "GET", Path: ""}:               {mediaEndpoint.List},
			{Method: "GET", Path: "/:mediaId"}:      {middlewares.PathParamIdMiddleware("mediaId"), mediaEndpoint.Get},
			{Method: "HEAD", Path: "/:mediaId"}:     {middlewares.PathParamIdMiddleware("mediaId"), mediaEndpoint.Get},
			{Method: "GET", Path: "/:mediaId/meta"}: {middlewares.PathParamIdMiddleware("mediaId"), mediaEndpoint.GetMetaData},
			{Method: "DELETE", Path: "/:mediaId"}:   {middlewares.PathParamIdMiddleware("mediaId"), mediaEndpoint.Delete},
			// Handle media chunk uploads with TUS to support huge file upload
			{Method: "POST", Path: "/chunkupload"}:             {gin.WrapH(http.StripPrefix("/media/chunkupload", http.HandlerFunc(handler.PostFile)))},
			{Method: "HEAD", Path: "/chunkupload/:uploadId"}:   {gin.WrapH(http.StripPrefix("/media/chunkupload", http.HandlerFunc(handler.HeadFile)))},
			{Method: "PATCH", Path: "/chunkupload/:uploadId"}:  {gin.WrapH(http.StripPrefix("/media/chunkupload", http.HandlerFunc(handler.PatchFile)))},
			{Method: "DELETE", Path: "/chunkupload/:uploadId"}: {gin.WrapH(http.StripPrefix("/media/chunkupload", http.HandlerFunc(handler.DelFile)))},
		},
		permissionsManager,
	)

	mediaEndpoint.EndpointGroup = endpoint
	return &mediaEndpoint
}

func (e *mediaEndpoint) PreCreate(hook tusd.HookEvent) (tusd.HTTPResponse, tusd.FileInfoChanges, error) {
	// Retrieve user that is uploading a file
	user, sharedLink, err := utils.GetUserOrSharedLinkGeneric(hook.Context)

	abortUnauthorized := func() (tusd.HTTPResponse, tusd.FileInfoChanges, error) {
		return tusd.HTTPResponse{}, handler.FileInfoChanges{}, handler.NewError("401", "Unauthorized", http.StatusUnauthorized)
	}

	if err != nil {
		return abortUnauthorized()
	}

	originalFilename := hook.Upload.MetaData["filename"]
	originalFilename, err = utils.ToUTF8(originalFilename)

	if err != nil || len(originalFilename) == 0 {
		return tusd.HTTPResponse{}, handler.FileInfoChanges{}, handler.NewError("400", "Bad request", http.StatusBadRequest)
	}

	mimeType := hook.Upload.MetaData["filetype"]
	fileExtension, err := utils.MimeTypeToFileExtension(mimeType)
	if err != nil {
		return tusd.HTTPResponse{}, handler.FileInfoChanges{}, handler.ErrInvalidContentType
	}

	// Check permissions
	if !e.GetPermissionsManager().CanCreateMedia(user, sharedLink) {
		return abortUnauthorized()
	}
	storageFileName := uuid.NewString() + "." + fileExtension
	dataDir, err := utils.GetDataDir("originalMedias")
	if err != nil {
		return tusd.HTTPResponse{}, handler.FileInfoChanges{}, handler.NewError("400", "Bad request", http.StatusBadRequest)
	}

	changes := handler.FileInfoChanges{
		Storage: map[string]string{
			"Path": filepath.Join(dataDir, storageFileName),
		},
		MetaData: map[string]string{
			"filename":         storageFileName,
			"originalFilename": originalFilename,
		},
	}

	if user != nil {
		changes.MetaData["userId"] = user.Id.Hex()
	}

	if sharedLink != nil {
		changes.MetaData["sharedLinkId"] = sharedLink.Id.Hex()
	}

	return handler.HTTPResponse{}, changes, nil
}

func (e *mediaEndpoint) PreFinish(hook handler.HookEvent) (handler.HTTPResponse, error) {
	// Remove the .info file
	dataDir, err := utils.GetDataDir("originalMedias")
	if err == nil {
		if err := os.Remove(filepath.Join(dataDir, hook.Upload.ID+".info")); err != nil {
			slog.Error("couldn't remove .info file", "upload ID", hook.Upload.ID, "error", err)
		}
	}
	hexUserId := hook.Upload.MetaData["userId"]
	hexSharedLinkId := hook.Upload.MetaData["sharedLinkId"]
	userId, errUserId := primitive.ObjectIDFromHex(hexUserId)
	_, errSharedId := primitive.ObjectIDFromHex(hexSharedLinkId)

	var userIdPtr *primitive.ObjectID = nil
	if errUserId == nil {
		userIdPtr = &userId
	}

	mediaId, svcErr := e.mediaService.Create(hook.Upload.MetaData["originalFilename"], hook.Upload.MetaData["filename"], userIdPtr, errSharedId != nil)
	if svcErr != nil {
		return tusd.HTTPResponse{}, handler.NewError(svcErr.GetMessage(), svcErr.GetMessage(), svcErr.GetCode())
	}
	return handler.HTTPResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf(`{"mediaId": "%s"}`, mediaId.Hex()),
		Header: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
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
	mimeType, mediaFile, modTime, svcErr := e.mediaService.GetData(*media.StorageFileName, compressedQuality)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}
	c.Header("Content-Type", *mimeType)

	if c.Request.Method == "GET" {
		http.ServeContent(c.Writer, c.Request, "", *modTime, mediaFile)
	}
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
