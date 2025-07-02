package utils_test

import (
	"data-storage-svc/internal/utils"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGetMimeType(t *testing.T) {
	testCases := []struct {
		filename          string
		expectedMimeType  string
		expectedExtension string
		isValid           bool
	}{
		{"testdata/sunflower.jpg", "image/jpeg", "jpg", true},
		{"testdata/gif.gif", "image/gif", "gif", false},
		{"testdata/video.mp4", "video/mp4", "mp4", true},
		{"testdata/classic-car.heic", "application/octet-stream", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			header, _ := utils.GetFileHeader(tc.filename)
			mimeType, extension, isValid := utils.CheckFileExtension(header)

			if isValid != tc.isValid {
				t.Errorf("Expected isValid %v, got %v", tc.isValid, isValid)
			}
			if mimeType != tc.expectedMimeType {
				t.Errorf("Expected MIME type %s, got %s", tc.expectedMimeType, mimeType)
			}
			if extension != tc.expectedExtension {
				t.Errorf("Expected file extension %s, got %s", tc.expectedExtension, extension)
			}
		})
	}
}

func TestCompress(t *testing.T) {
	resultDir := t.TempDir()
	testCases := []struct {
		originalFile     string
		destinationFile  string
		compressionLevel int
		exprectError     bool
	}{
		{"testdata/video.mp4", filepath.Join(resultDir, "mp4_compressed.mp4"), 23, false},
		{"testdata/sunflower.jpg", filepath.Join(resultDir, "jpg_compressed.jpg"), 23, false},
		{"testdata/sunflower.jpg", filepath.Join(resultDir, "jpg35_compressed.jpg"), 35, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%d", tc.originalFile, tc.compressionLevel), func(t *testing.T) {
			err := utils.CompressMedia(tc.originalFile, tc.destinationFile, tc.compressionLevel)
			if (err != nil) != tc.exprectError {
				t.Errorf("Expected error: %v, got: %v", tc.exprectError, err)
			}
			if err == nil {
				resultInfo, _ := os.Stat(tc.destinationFile)
				originalInfo, _ := os.Stat(tc.originalFile)
				if resultInfo.Size() == 0 {
					t.Errorf("Expected non-empty file: %s", tc.destinationFile)
				} else if resultInfo.Size() > originalInfo.Size() {
					t.Errorf("Expected compressed file to be smaller than original, but got %d bytes vs %d bytes",
						resultInfo.Size(), originalInfo.Size())
				}
			}
		})
	}
}
