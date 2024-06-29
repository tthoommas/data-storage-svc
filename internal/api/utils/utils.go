package utils

import (
	"data-storage-svc/internal"
	"log/slog"
	"os"
	"path/filepath"
)

func GetDataDir(subPath string) (string, error) {
	path := filepath.Join(internal.DATA_DIR_PATH, subPath)
	exists, err := pathExists(path)
	if err != nil || !exists {
		slog.Debug("Data path not found", "path", path)
		return "", err
	}
	return path, nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil // path exists
	}
	if os.IsNotExist(err) {
		return false, nil // path does not exist
	}
	return false, err // some other error
}
