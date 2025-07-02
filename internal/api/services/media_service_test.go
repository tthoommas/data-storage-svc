package services_test

import (
	"data-storage-svc/internal"
	"data-storage-svc/internal/api/services"
	"data-storage-svc/internal/mocks"
	"data-storage-svc/internal/model"
	"data-storage-svc/internal/utils"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCreate(t *testing.T) {

	testCases := []struct {
		name                string
		filename            string
		uploader            primitive.ObjectID
		uploadViaSharedLink bool
		data                io.ReadCloser
		expectedErrorCode   *int
	}{
		{
			name:                "Invalid filename",
			uploader:            primitive.NewObjectID(),
			filename:            "",
			uploadViaSharedLink: false,
			expectedErrorCode:   utils.IntPtr(400),
		},
		{
			name:                "No data provided",
			uploader:            primitive.NewObjectID(),
			filename:            "myfile.jpg",
			data:                nil,
			uploadViaSharedLink: false,
			expectedErrorCode:   utils.IntPtr(400),
		},
		{
			name:                "Unsupported extension",
			uploader:            primitive.NewObjectID(),
			filename:            "myfile.exe",
			data:                getData("cat.jpg"),
			uploadViaSharedLink: false,
			expectedErrorCode:   utils.IntPtr(400),
		},
		{
			name:                "Invalid file name",
			uploader:            primitive.NewObjectID(),
			filename:            "toto",
			data:                getData("cat.jpg"),
			uploadViaSharedLink: false,
			expectedErrorCode:   utils.IntPtr(400),
		},
		{
			name:                "Create JPG success",
			uploader:            ObjIdFromHex("67fbd784c491ff384ee6287d"),
			filename:            "cat.jpg",
			data:                getData("cat.jpg"),
			uploadViaSharedLink: false,
			expectedErrorCode:   nil,
		},
		{
			name:                "Create PNG success",
			uploader:            ObjIdFromHex("67fbe42de0d2f5f686c2127c"),
			filename:            "transparent.png",
			data:                getData("transparent.png"),
			uploadViaSharedLink: false,
			expectedErrorCode:   nil,
		},
	}

	internal.DATA_DIRECTORY = t.TempDir()

	mediaRepositoryMock := mocks.MediaRepository{}
	// Expect create call for jpg
	mediaRepositoryMock.On("Create", mock.MatchedBy(func(media *model.Media) bool {
		if *media.OriginalFileName != "cat.jpg" {
			return false
		}
		if media.UploadedBy.Hex() != "67fbd784c491ff384ee6287d" {
			return false
		}
		if time.Since(*media.UploadTime).Seconds() > 2 {
			return false
		}
		return true
	})).Return(utils.Ptr(primitive.NewObjectID()), nil).Once()

	// Expect create call for png
	mediaRepositoryMock.On("Create", mock.MatchedBy(func(media *model.Media) bool {
		if *media.OriginalFileName != "transparent.png" {
			return false
		}
		if media.UploadedBy.Hex() != "67fbe42de0d2f5f686c2127c" {
			return false
		}
		if time.Since(*media.UploadTime).Seconds() > 2 {
			return false
		}
		return true
	})).Return(utils.Ptr(primitive.NewObjectID()), nil).Once()

	mediaInAlbumRepositoryMock := mocks.MediaInAlbumRepository{}
	mediaAccessServiceMock := mocks.MediaAccessService{}
	albumServiceMock := mocks.AlbumService{}

	mediaService := services.NewMediaService(&mediaRepositoryMock, &mediaInAlbumRepositoryMock, &mediaAccessServiceMock, &albumServiceMock)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createdId, err := mediaService.Create(tc.filename, &tc.uploader, tc.uploadViaSharedLink, &tc.data)
			if tc.expectedErrorCode != nil {
				assert.Nil(t, createdId)
				assert.NotNil(t, err)
				assert.Equal(t, *tc.expectedErrorCode, err.GetCode())
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, createdId)
			}
		})
	}
	mediaRepositoryMock.AssertExpectations(t)

	originals, _ := utils.GetDataDir("originalMedias")
	originalsFiles, _ := os.ReadDir(originals)
	originalFileNames := []string{}
	originalFileExtensions := []string{}

	for _, f := range originalsFiles {
		r := strings.Split(f.Name(), ".")
		originalFileNames = append(originalFileNames, r[0])
		originalFileExtensions = append(originalFileExtensions, r[1])
	}

	compressed, _ := utils.GetDataDir("compressedMedias")
	compressedFiles, _ := os.ReadDir(compressed)
	compressedFileNames := []string{}
	for _, f := range compressedFiles {
		r := strings.Split(f.Name(), ".")
		compressedFileNames = append(compressedFileNames, r[0])
		// Assert we only have jpg in the compressed foleder, whatever is the original extension
		assert.Equal(t, "jpg", r[1])
	}

	// Assert we have exactly one compressed file for each original file with the same name
	assert.Equal(t, originalFileNames, compressedFileNames)
	// Assert file extensions are preserved for original files
	assert.ElementsMatch(t, []string{"jpg", "png"}, originalFileExtensions)
}

func TestMemory(t *testing.T) {
	internal.DATA_DIRECTORY = t.TempDir()
	mediaRepositoryMock := mocks.MediaRepository{}
	mediaRepositoryMock.On("Create", mock.Anything).Return(utils.Ptr(primitive.NewObjectID()), nil)

	mediaInAlbumRepositoryMock := mocks.MediaInAlbumRepository{}
	mediaAccessServiceMock := mocks.MediaAccessService{}
	albumServiceMock := mocks.AlbumService{}

	mediaService := services.NewMediaService(&mediaRepositoryMock, &mediaInAlbumRepositoryMock, &mediaAccessServiceMock, &albumServiceMock)

	uploader := primitive.NewObjectID()

	f, err := os.Create("test_cpu.prof")
	if err != nil {
		t.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer func() {
		pprof.StopCPUProfile()
		f.Close()
	}()

	for range 100 {
		data := getData("big_image.jpg")
		newId, err := mediaService.Create("newFile.jpg", &uploader, false, &data)
		assert.NotNil(t, newId)
		assert.Nil(t, err)
	}
}

func getData(filename string) io.ReadCloser {
	file, err := os.Open(filepath.Join("testdata", filename))
	if err != nil {
		panic(err)
	}
	return file
}

func ObjIdFromHex(hex string) primitive.ObjectID {
	r, _ := primitive.ObjectIDFromHex(hex)
	return r
}
