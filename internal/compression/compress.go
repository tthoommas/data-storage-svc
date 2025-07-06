package compression

import (
	"data-storage-svc/internal/utils"
	"fmt"
	"log/slog"
	"os/exec"
)

var isFfmpegInstalled *bool = nil

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

	h, err := utils.GetFileHeader(originalFilePath)
	if err != nil {
		return err
	}
	mimeType, _, _ := utils.CheckFileExtension(h)

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
