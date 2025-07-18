package compression

import (
	"data-storage-svc/internal/utils"
	"fmt"
	"os/exec"
	"path/filepath"
)

var isFfmpegInstalled *bool = nil

// Compress a unique media using ffmpeg
// Supported input format
// - jpeg/jpg
// - png
// - mp4
// On success, returns the name of the compressed file version in the destination folder
func CompressMedia(originalFilePath, destinationFolder string) (*string, error) {
	if isFfmpegInstalled == nil {
		_, err := exec.LookPath("ffmpeg")
		isFfmpegInstalled = &[]bool{err == nil}[0]
	}
	if !*isFfmpegInstalled {
		panic("ffmpeg is not installed on this system, cannot compress media")
	}

	h, err := utils.GetFileHeader(originalFilePath)
	if err != nil {
		return nil, err
	}
	mimeType, _, _ := utils.CheckFileExtension(h)

	switch mimeType {
	case "image/jpeg", "image/png", "image/heic":
		return compressImage(originalFilePath, destinationFolder)
	case "video/mp4":
		return compressVideo(originalFilePath, destinationFolder)
	default:
		return nil, fmt.Errorf("unsupported media type: %s", mimeType)
	}
}

func compressImage(originalFilePath string, destinationFolder string) (*string, error) {
	// Compress all image files as jpg whatever is the original format
	compressFilename := filepath.Base(originalFilePath) + ".jpg"
	cmd := exec.Command("ffmpeg",
		"-i", originalFilePath,
		"-vf", "scale=720:-1",
		"-q:v", "6",
		"-y",
		filepath.Join(destinationFolder, compressFilename))
	_, err := cmd.CombinedOutput()
	return &compressFilename, err
}

func compressVideo(originalFilaPath string, destinationFolder string) (*string, error) {
	// First try hardware optimizations
	if compressFilename, err := compressMP4HW(originalFilaPath, destinationFolder); err == nil {
		return compressFilename, nil
	}
	// Fallback to software compression
	return compressMP4Soft(originalFilaPath, destinationFolder)
}

// Compress MP4 with raspberry pi HW acceleration
func compressMP4HW(originalFilePath string, destinationFolder string) (*string, error) {
	compressFilename := filepath.Base(originalFilePath)
	cmd := exec.Command("ffmpeg",
		"-i", originalFilePath,
		"-vf", "scale='if(gt(iw,1280),1280,trunc(iw/16)*16)':'if(gt(ih,720),720,trunc(ih/16)*16)',fps=30",
		"-c:v", "h264_v4l2m2m",
		"-b:v", "4M",
		"-c:a", "copy",
		"-y",
		filepath.Join(destinationFolder, compressFilename))
	_, err := cmd.CombinedOutput()
	return &compressFilename, err
}

func compressMP4Soft(originalFilePath string, destinationFolder string) (*string, error) {
	compressFilename := filepath.Base(originalFilePath)
	cmd := exec.Command("ffmpeg",
		"-i", originalFilePath,
		"-vf", "scale='if(gt(iw,1280),1280,trunc(iw/16)*16)':'if(gt(ih,720),720,trunc(ih/16)*16)',fps=30",
		"-c:v", "libx264",
		"-b:v", "4M",
		"-preset", "ultrafast",
		"-c:a", "copy",
		"-y",
		filepath.Join(destinationFolder, compressFilename))
	_, err := cmd.CombinedOutput()
	return &compressFilename, err
}
