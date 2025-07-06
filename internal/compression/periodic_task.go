package compression

import (
	"log/slog"
	"time"
)

func PeriodCompression(delaySeconds int) {
	for {
		filesToCompress, err := getFilesToCompress()
		if err != nil {
			slog.Error("Failed to get files to compress in goroutine, skipping compression", "err", err)
			wait(delaySeconds)
			continue
		}
		err = compressFiles(filesToCompress)
		if err != nil {
			slog.Error("Failed to compress files", "error", err)
		}
		wait(delaySeconds)
	}
}

func wait(delaySeconds int) {
	time.Sleep(time.Duration(delaySeconds) * time.Second)
}

func getFilesToCompress() ([]string, error) {
	return nil, nil
}

func compressFiles(filesToCompress []string) error {
	return nil
}
