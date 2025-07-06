package compression_test

import (
	"data-storage-svc/internal/compression"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCompress(t *testing.T) {
	resultDir := t.TempDir()
	testCases := []struct {
		originalFile     string
		destinationFile  string
		compressionLevel int
		exprectError     bool
	}{
		{"../testdata/video.mp4", filepath.Join(resultDir, "mp4_compressed.mp4"), 23, false},
		{"../testdata/sunflower.jpg", filepath.Join(resultDir, "jpg_compressed.jpg"), 23, false},
		{"../testdata/sunflower.jpg", filepath.Join(resultDir, "jpg35_compressed.jpg"), 35, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%d", tc.originalFile, tc.compressionLevel), func(t *testing.T) {
			err := compression.CompressMedia(tc.originalFile, tc.destinationFile, tc.compressionLevel)
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
