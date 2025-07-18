package compression

import (
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/repository"
	"data-storage-svc/internal/utils"
	"log/slog"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Compression queue, containing file names to compress
var compressionQueue = make(chan primitive.ObjectID, 15)

// Periodic task to compress files after they have been uploaded. It will run every delaySeconds, check if there are files
// that have not been compressed yet, and if yes, compress them.
func CompressionTask(delaySeconds int64, mediaRepository repository.MediaRepository) {
	originalDir, errOri := utils.GetDataDir(common.ORIGINAL_MEDIA_DIRECTORY)
	if errOri != nil {
		slog.Error("couldn't open original media directory", "error", errOri)
		return
	}
	compressedDir, errComp := utils.GetDataDir(common.COMPRESSED_DIRECTORY)
	if errComp != nil {
		slog.Error("couldn't open compressed media directory", "error", errComp)
		return
	}
	ticker := time.NewTicker(time.Duration(delaySeconds) * time.Second)
	for range ticker.C {
		slog.Debug("Executing compression task")
		compressBatch := map[primitive.ObjectID]bool{}
		// Collect files to compress from the channel
		collectChannel := true
	CollectFromChannel:
		for collectChannel {
			// Compress 15 files by batch at maximum
			if len(compressBatch) >= 15 {
				slog.Debug("Compress batch is full, proceeding to compression")
				break CollectFromChannel
			}
			select {
			case mediaIdToCompress := <-compressionQueue:
				compressBatch[mediaIdToCompress] = true
			default:
				slog.Debug("Proceeding to compression")
				break CollectFromChannel
			}
		}
		nbrCompressedFile := 0
		nbrFailed := 0
		// Now compress the files
		for mediaId := range compressBatch {
			media, err := mediaRepository.Get(&mediaId)
			if err != nil {
				slog.Error("Couldn't fetch media to compress, skipping", "error", err)
				continue
			}
			if name, err := CompressMedia(filepath.Join(originalDir, *media.StorageFileName), compressedDir); err != nil {
				slog.Error("Couldn't compress file", "filename", *media.OriginalFileName, "error", err)
				nbrFailed += 1
			} else {
				nbrCompressedFile += 1
				setCompressedFileName(&mediaId, name, mediaRepository)
			}
		}
		slog.Debug("Compression results", "successes", nbrCompressedFile, "failures", nbrFailed)
	}
}

func setCompressedFileName(mediaId *primitive.ObjectID, compressedFileName *string, mediaRepository repository.MediaRepository) {
	update := bson.M{}
	if compressedFileName != nil {
		update["compressedFileName"] = *compressedFileName
	} else {
		update["compressedFileName"] = nil
	}
	if err := mediaRepository.Update(mediaId, update); err != nil {
		slog.Error("error setting compressed file name", "error", err)
	}
}

func AddToCompressQueue(mediaId *primitive.ObjectID) {
	if mediaId != nil {
		compressionQueue <- *mediaId
	}
}
