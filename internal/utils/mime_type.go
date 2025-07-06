package utils

import (
	"net/http"
	"os"
	"slices"
)

var ACCEPTED_FILE_EXTENSIONS = []string{"jpg", "jpeg", "png", "mp4"}

func CheckFileExtension(fileHeader []byte) (string, string, bool) {
	mimeType := http.DetectContentType(fileHeader)
	extension := mimeTypeToFileExtension(mimeType)
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

func mimeTypeToFileExtension(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "video/mp4":
		return "mp4"
	case "image/heic":
		return "heic"
	case "image/gif":
		return "gif"
	default:
		return ""
	}
}
