package utils_test

import (
	"data-storage-svc/internal/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashMedia(t *testing.T) {
	testCase := []struct {
		name          string
		inputFile     string
		expectedError string
		expectedHash  *string
	}{
		{
			name:          "Hash success",
			inputFile:     "../testdata/video.mp4",
			expectedError: "",
			expectedHash:  utils.StrPtr("71944d7430c461f0cd6e7fd10cee7eb72786352a3678fc7bc0ae3d410f72aece"),
		},
		{
			name:          "Hash success 2",
			inputFile:     "../testdata/sunflower.jpg",
			expectedError: "",
			expectedHash:  utils.StrPtr("d488eff532dd2c9267578acf815ac2a20eb18a321ab4a6a86ef2253312358624"),
		},
		{
			name:          "Unexisting file",
			inputFile:     "../testdata/sunflower2.jpg",
			expectedError: "open ../testdata/sunflower2.jpg: no such file or directory",
			expectedHash:  nil,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			hashResult, err := utils.HashMedia(tc.inputFile)
			assert.Equal(t, tc.expectedError == "", err == nil)
			if err != nil {
				assert.Equal(t, tc.expectedError, err.Error())
			}
			assert.Equal(t, tc.expectedHash == nil, hashResult == nil)
			if tc.expectedHash != nil {
				assert.Equal(t, *tc.expectedHash, *hashResult)
			}
		})
	}
}
