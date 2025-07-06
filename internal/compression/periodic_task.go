package compression

import (
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/utils"
	"log/slog"
	"path/filepath"
	"time"
)

// Compression queue, containing file names to compress
var compressionQueue = make(chan string, 100)

// Periodic task to compress files after they have been uploaded. It will run every delaySeconds, check if there are files
// that have not been compressed yet, and if yes, compress them.
func CompressionTask(delaySeconds int64) {
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
		compressBatch := map[string]bool{}
		// Collect files to compress from the channel
		collectChannel := true
	CollectFromChannel:
		for collectChannel {
			// Compress 100 files by batch at maximum
			if len(compressBatch) >= 100 {
				slog.Debug("Compress batch is full, proceeding to compression")
				break CollectFromChannel
			}
			select {
			case fileToCompress := <-compressionQueue:
				compressBatch[fileToCompress] = true
			default:
				slog.Debug("Proceeding to compression")
				break CollectFromChannel
			}
		}
		nbrCompressedFile := 0
		nbrFailed := 0
		// Now compress the files
		for filename := range compressBatch {
			if err := CompressMedia(filepath.Join(originalDir, filename), compressedDir); err != nil {
				slog.Error("Couldn't compress file", "filename", filename, "error", err)
				nbrFailed += 1
			} else {
				nbrCompressedFile += 1
			}
		}
		slog.Debug("Compression results", "successes", nbrCompressedFile, "failures", nbrFailed)
	}
}

func AddToCompressQueue(filename string) {
	compressionQueue <- filename
}
