package utils_test

import (
	"data-storage-svc/internal/utils"
	"testing"
)

func TestGetMimeType(t *testing.T) {
	testCases := []struct {
		filename          string
		expectedMimeType  string
		expectedExtension string
		isValid           bool
	}{
		{"../testdata/sunflower.jpg", "image/jpeg", "jpg", true},
		{"../testdata/gif.gif", "image/gif", "gif", false},
		{"../testdata/video.mp4", "video/mp4", "mp4", true},
		{"../testdata/classic-car.heic", "application/octet-stream", "", false},
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
