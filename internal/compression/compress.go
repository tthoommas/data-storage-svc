package compression

import (
	"data-storage-svc/internal/utils"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
)

var isFfmpegInstalled *bool = nil

// Compress a unique media using ffmpeg
// Supported input format
// - jpeg/jpg
// - png
// - mp4
func CompressMedia(originalFilePath, destinationFolder string) error {
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

	var compressCmd []exec.Cmd
	switch mimeType {
	case "image/jpeg", "image/png", "image/heic":
		compressCmd = append(compressCmd, *compressImage(originalFilePath, destinationFolder))
	case "video/mp4":
		// By default use hardware acceleration
		compressCmd = append(compressCmd, *compressMP4HW(originalFilePath, destinationFolder))
		// Fallback to software compression
		compressCmd = append(compressCmd, *compressMP4Soft(originalFilePath, destinationFolder))
	default:
		return fmt.Errorf("unsupported media type: %s", mimeType)
	}

	errors := []error{}
	outputs := [][]byte{}
	for _, cmd := range compressCmd {
		output, err := cmd.CombinedOutput()
		if err == nil {
			// Compression worked, end here
			return nil
		} else {
			errors = append(errors, err)
			outputs = append(outputs, output)
		}
	}
	// All compression methods tried, nothing worked
	for i, err := range errors {
		slog.Error("Failed to compress media", "error", err, "output", string(outputs[i]))
	}
	return fmt.Errorf("couldn't compress media")
}

func compressImage(originalFilePath string, destinationFolder string) *exec.Cmd {
	// Compress all image files as jpg whatever is the original format
	originalFile := filepath.Base(originalFilePath)
	return exec.Command("ffmpeg",
		"-i", originalFilePath,
		"-vf", "scale=720:-1",
		"-q:v", "6",
		"-y",
		filepath.Join(destinationFolder, originalFile))
}

// Compress MP4 with raspberry pi HW acceleration
func compressMP4HW(originalFilePath string, destinationFolder string) *exec.Cmd {
	originalFileName := filepath.Base(originalFilePath)
	return exec.Command("ffmpeg",
		"-i", originalFilePath,
		"-vf", "scale='if(gt(iw,1280),1280,trunc(iw/16)*16)':'if(gt(ih,720),720,trunc(ih/16)*16)',fps=30",
		"-c:v", "h264_v4l2m2m",
		"-b:v", "4M",
		"-c:a", "copy",
		"-y",
		filepath.Join(destinationFolder, originalFileName))
}

func compressMP4Soft(originalFilePath string, destinationFolder string) *exec.Cmd {
	originalFileName := filepath.Base(originalFilePath)
	return exec.Command("ffmpeg",
		"-i", originalFilePath,
		"-vf", "scale='if(gt(iw,1280),1280,trunc(iw/16)*16)':'if(gt(ih,720),720,trunc(ih/16)*16)',fps=30",
		"-c:v", "libx264",
		"-b:v", "4M",
		"-preset", "ultrafast",
		"-c:a", "copy",
		"-y",
		filepath.Join(destinationFolder, originalFileName))
}
