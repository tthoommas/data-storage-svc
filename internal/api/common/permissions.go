package common

import (
	"data-storage-svc/internal/model"
	"data-storage-svc/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PermissionsManager interface {
	CanListUsers(user *model.User) bool
	CanCreateUser(user *model.User) bool
	CanCreateAlbum(user *model.User) bool
	CanGetAlbum(user *model.User, albumId *primitive.ObjectID, sharedLink *model.SharedLink) bool
	CanGetAllMediasForAlbum(user *model.User, albumId *primitive.ObjectID, sharedLink *model.SharedLink) bool
	CanEditMediasInAlbum(user *model.User, albumId *primitive.ObjectID, sharedLink *model.SharedLink) bool
	CanDeleteAlbum(user *model.User, albumId *primitive.ObjectID) bool
	CanListAlbumAccesses(user *model.User, albumId *primitive.ObjectID) bool
	CanEditAlbumAccesses(user *model.User, albumId *primitive.ObjectID) bool
	CanInitDownloadForAlbum(user *model.User, albumId *primitive.ObjectID, sharedLink *model.SharedLink) bool
	CanConsumeDownload(user *model.User, downloadId *primitive.ObjectID, sharedLink *model.SharedLink) bool
	CanGetDownload(user *model.User, downloadId *primitive.ObjectID, sharedLink *model.SharedLink) bool
	CanCreateMedia(user *model.User, sharedLink *model.SharedLink) bool
	CanGetMedia(user *model.User, mediaId *primitive.ObjectID, sharedLink *model.SharedLink) bool
	CanDeleteMedia(user *model.User, mediaId *primitive.ObjectID) bool
	CanCreateSharedLink(user *model.User, albumId *primitive.ObjectID) bool
	CanListSharedLinks(user *model.User, albumId *primitive.ObjectID) bool
	CanDeleteSharedLink(user *model.User, sharedLink *model.SharedLink) bool
	CanUpdateSharedLink(user *model.User, sharedLink *model.SharedLink) bool
}

type permissionsManager struct {
	albumAccessRepository  repository.AlbumAccessRepository
	albumRepository        repository.AlbumRepository
	downloadRepository     repository.DownloadRepository
	mediaAccessRepository  repository.MediaAccessRepository
	mediaInAblumRepository repository.MediaInAlbumRepository
	mediaRepository        repository.MediaRepository
}

func NewPermissionsManager(albumAccessRepository repository.AlbumAccessRepository, albumRepository repository.AlbumRepository, downloadRepository repository.DownloadRepository, mediaAccessRepository repository.MediaAccessRepository, mediaInAblumRepository repository.MediaInAlbumRepository, mediaRepository repository.MediaRepository) PermissionsManager {
	return permissionsManager{albumAccessRepository: albumAccessRepository, albumRepository: albumRepository, downloadRepository: downloadRepository, mediaAccessRepository: mediaAccessRepository, mediaInAblumRepository: mediaInAblumRepository, mediaRepository: mediaRepository}
}

func (p permissionsManager) CanListUsers(user *model.User) bool {
	return user != nil && user.IsAdmin
}

func (p permissionsManager) CanCreateUser(user *model.User) bool {
	return user != nil && user.IsAdmin
}

func (p permissionsManager) CanCreateAlbum(user *model.User) bool {
	return user != nil
}

func (p permissionsManager) CanGetAlbum(user *model.User, albumId *primitive.ObjectID, sharedLink *model.SharedLink) bool {
	return p.getAlbumAccessOrNil(user, albumId) != nil || (sharedLink != nil && sharedLink.AlbumId.Hex() == albumId.Hex())
}

func (p permissionsManager) CanGetAllMediasForAlbum(user *model.User, albumId *primitive.ObjectID, sharedLink *model.SharedLink) bool {
	return p.CanGetAlbum(user, albumId, sharedLink)
}

func (p permissionsManager) CanEditMediasInAlbum(user *model.User, albumId *primitive.ObjectID, sharedLink *model.SharedLink) bool {
	access := p.getAlbumAccessOrNil(user, albumId)
	return (access != nil && access.CanEdit) || (sharedLink != nil && sharedLink.AlbumId.Hex() == albumId.Hex() && sharedLink.CanEdit)
}

func (p permissionsManager) CanDeleteAlbum(user *model.User, albumId *primitive.ObjectID) bool {
	return p.isAlbumAuthor(user, albumId)
}

func (p permissionsManager) CanListAlbumAccesses(user *model.User, albumId *primitive.ObjectID) bool {
	return p.isAlbumAuthor(user, albumId)
}

func (p permissionsManager) CanEditAlbumAccesses(user *model.User, albumId *primitive.ObjectID) bool {
	return p.isAlbumAuthor(user, albumId)
}

func (p permissionsManager) CanInitDownloadForAlbum(user *model.User, albumId *primitive.ObjectID, sharedLink *model.SharedLink) bool {
	return p.CanGetAlbum(user, albumId, sharedLink)
}

func (p permissionsManager) CanConsumeDownload(user *model.User, downloadId *primitive.ObjectID, sharedLink *model.SharedLink) bool {
	return p.isDownloadAuthor(user, downloadId, sharedLink)
}

func (p permissionsManager) CanGetDownload(user *model.User, downloadId *primitive.ObjectID, sharedLink *model.SharedLink) bool {
	return p.isDownloadAuthor(user, downloadId, sharedLink)
}

func (p permissionsManager) CanCreateMedia(user *model.User, sharedLink *model.SharedLink) bool {
	return user != nil || (sharedLink != nil && sharedLink.CanEdit)
}

func (p permissionsManager) CanGetMedia(user *model.User, mediaId *primitive.ObjectID, sharedLink *model.SharedLink) bool {
	if p.isMediaAuthor(user, mediaId) {
		// If user owns this media, directly grant acccess
		return true

	} else if sharedLink != nil && p.mediaInAblumRepository.IsInAlbum(mediaId, &sharedLink.AlbumId) {
		// Otherwise check if user is accessing this media via a shared link
		return true
	} else {
		// Finally check if this media belongs to an album that user is allowed to view
		userAlbums, err := p.albumAccessRepository.GetAllByUser(&user.Id)
		isInUsersAlbums := false
		if err == nil {
			for _, album := range userAlbums {
				if p.mediaInAblumRepository.IsInAlbum(mediaId, album.AlbumId) {
					isInUsersAlbums = true
					break
				}
			}
		}
		return isInUsersAlbums
	}
}

func (p permissionsManager) CanDeleteMedia(user *model.User, mediaId *primitive.ObjectID) bool {
	return p.isMediaAuthor(user, mediaId)
}

func (p permissionsManager) CanCreateSharedLink(user *model.User, albumId *primitive.ObjectID) bool {
	return p.isAlbumAuthor(user, albumId)
}

func (p permissionsManager) CanListSharedLinks(user *model.User, albumId *primitive.ObjectID) bool {
	return p.isAlbumAuthor(user, albumId)
}

func (p permissionsManager) CanDeleteSharedLink(user *model.User, sharedLink *model.SharedLink) bool {
	return user != nil && sharedLink != nil && sharedLink.CreatedBy.Hex() == user.Id.Hex()
}

func (p permissionsManager) CanUpdateSharedLink(user *model.User, sharedLink *model.SharedLink) bool {
	return user != nil && sharedLink != nil && sharedLink.CreatedBy.Hex() == user.Id.Hex()
}

// Utility private methods
func (p permissionsManager) isMediaAuthor(user *model.User, mediaId *primitive.ObjectID) bool {
	media := p.getMediaOrNil(user, mediaId)
	return media != nil && media.UploadedBy.Hex() == user.Id.Hex()
}

func (p permissionsManager) getMediaOrNil(user *model.User, mediaId *primitive.ObjectID) *model.Media {
	if user == nil || mediaId == nil {
		return nil
	}
	media, _ := p.mediaRepository.Get(mediaId)
	return media
}

func (p permissionsManager) getMediaAccessOrNil(user *model.User, mediaId *primitive.ObjectID) *model.UserMediaAccess {
	if user == nil || mediaId == nil {
		return nil
	}
	userMediaAccess, _ := p.mediaAccessRepository.Get(&user.Id, mediaId)
	return userMediaAccess
}

func (p permissionsManager) getAlbumAccessOrNil(user *model.User, albumId *primitive.ObjectID) *model.UserAlbumAccess {
	if user == nil || albumId == nil {
		return nil
	}
	userAlbumAccess, _ := p.albumAccessRepository.Get(&user.Id, albumId)
	return userAlbumAccess
}

func (p permissionsManager) getAlbum(albumId *primitive.ObjectID) *model.Album {
	if albumId == nil {
		return nil
	}
	album, _ := p.albumRepository.GetById(*albumId)
	return album
}

func (p permissionsManager) getDownload(downloadId *primitive.ObjectID) *model.Download {
	if downloadId == nil {
		return nil
	}
	download, _ := p.downloadRepository.Get(downloadId)
	return download
}

func (p permissionsManager) isAlbumAuthor(user *model.User, albumId *primitive.ObjectID) bool {
	album := p.getAlbum(albumId)
	return user != nil && album != nil && user.Id.Hex() == album.AuthorId.Hex()
}

func (p permissionsManager) isDownloadAuthor(user *model.User, downloadId *primitive.ObjectID, sharedLink *model.SharedLink) bool {
	download := p.getDownload(downloadId)
	return download != nil && download.Initiator != nil && ((user != nil && download.Initiator.Hex() == user.Id.Hex()) || (sharedLink != nil && download.Initiator.Hex() == sharedLink.Id.Hex()))
}
