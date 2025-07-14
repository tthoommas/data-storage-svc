package utils

import (
	"fmt"
	"net/http"
	"os"
	"slices"
)

var ACCEPTED_FILE_EXTENSIONS = []string{"jpg", "jpeg", "png", "mp4"}

func CheckFileExtension(fileHeader []byte) (string, string, bool) {
	mimeType := http.DetectContentType(fileHeader)
	extension, _ := MimeTypeToFileExtension(mimeType)
	return mimeType, extension, slices.Contains(ACCEPTED_FILE_EXTENSIONS, extension)
}

func GetFileHeader(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileHeader := make([]byte, 512)
	_, err = file.Read(fileHeader)
	if err != nil {
		return nil, err
	}

	return fileHeader, nil
}

func MimeTypeToFileExtension(mimeType string) (string, error) {
	switch mimeType {
	case "image/jpeg":
		return "jpg", nil
	case "image/png":
		return "png", nil
	case "video/mp4":
		return "mp4", nil
	case "image/heic":
		return "heic", nil
	case "image/gif":
		return "gif", nil
	default:
		return "", fmt.Errorf("unknown mime type")
	}
}
