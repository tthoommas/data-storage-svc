package compression_test

import (
	"data-storage-svc/internal/compression"
	"os"
	"testing"
)

func TestCompress(t *testing.T) {
	resultDir := t.TempDir()
	testCases := []struct {
		originalFile    string
		destinationFile string
		expectError     bool
	}{
		{"../testdata/video.mp4", resultDir, false},
		{"../testdata/sunflower.jpg", resultDir, false},
	}

	for _, tc := range testCases {
		t.Run(tc.originalFile, func(t *testing.T) {
			_, err := compression.CompressMedia(tc.originalFile, tc.destinationFile)
			if (err != nil) != tc.expectError {
				t.Errorf("Expected error: %v, got: %v", tc.expectError, err)
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
