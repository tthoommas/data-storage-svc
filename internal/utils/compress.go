package utils

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"slices"
)

var isFfmpegInstalled *bool = nil
var ACCEPTED_FILE_EXTENSIONS = []string{"jpg", "jpeg", "png", "mp4"}

func CompressMedia(originalFilePath, destinationFilePath string, compressionLevel int) error {
	if compressionLevel < 2 || compressionLevel > 32 {
		return fmt.Errorf("invalid compression level: %d, must be between 2 and 32", compressionLevel)
	}
	if isFfmpegInstalled == nil {
		_, err := exec.LookPath("ffmpeg")
		isFfmpegInstalled = &[]bool{err == nil}[0]
	}
	if !*isFfmpegInstalled {
		panic("ffmpeg is not installed on this system, cannot compress media")
	}

	h, err := GetFileHeader(originalFilePath)
	if err != nil {
		return err
	}
	mimeType, _, _ := CheckFileExtension(h)

	var compressCmd *exec.Cmd
	switch mimeType {
	case "image/jpeg", "image/png", "image/heic":
		compressCmd = exec.Command("ffmpeg", "-i", originalFilePath, "-q:v", fmt.Sprintf("%d", compressionLevel), destinationFilePath)
	case "video/mp4":
		compressCmd = exec.Command("ffmpeg", "-i", originalFilePath, "-vf", "scale=-2:720,fps=30", "-c:v", "libx264", "-preset", "ultrafast", "-crf", fmt.Sprintf("%d", compressionLevel), "-c:a", "aac", "-b:a", "96k", destinationFilePath)
	default:
		return fmt.Errorf("unsupported media type: %s", mimeType)
	}

	output, err := compressCmd.CombinedOutput()
	if err != nil {
		slog.Debug("Failed to compress media", "error", err, "output", string(output))
		return err
	}
	return nil
}

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
