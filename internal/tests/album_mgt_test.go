package tests

import (
	"data-storage-svc/internal/model"
	"fmt"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestCreateAlbum(t *testing.T) {
	cases := []struct {
		name                 string
		album                string
		client               *resty.Client
		expectError          bool
		expectedResponseCode int
	}{
		{
			name:                 "Fail unauthenticated",
			album:                `{"albumTitle": "title", "albumDescription": "description"}`,
			client:               clients[UNAUTHENTICATED_CLIENT].restyClient,
			expectError:          false,
			expectedResponseCode: 401,
		},
		{
			name:                 "Create success",
			album:                `{"albumTitle": "title", "albumDescription": "description"}`,
			client:               clients[USER1].restyClient,
			expectError:          false,
			expectedResponseCode: 201,
		},
		{
			name:                 "Fail no title",
			album:                `{"albumTitle": "", "albumDescription": "description"}`,
			client:               clients[USER1].restyClient,
			expectError:          false,
			expectedResponseCode: 400,
		},
		{
			name:                 "Success no description",
			album:                `{"albumTitle": "Title"}`,
			client:               clients[USER1].restyClient,
			expectError:          false,
			expectedResponseCode: 201,
		},
		{
			name:                 "Fail no description no title",
			album:                `{}`,
			client:               clients[USER1].restyClient,
			expectError:          false,
			expectedResponseCode: 400,
		},
	}

	for _, c := range cases {
		resp, err := c.client.R().SetBody(c.album).Post(TEST_API + "/album/create")
		if !c.expectError {
			assert.Nil(t, err)
			assert.Equal(t, c.expectedResponseCode, resp.StatusCode())
		} else {
			assert.NotNil(t, err)
		}
	}
}

func TestGetAlbumById(t *testing.T) {
	var respBody struct {
		AlbumId string `json:"albumId"`
	}
	clients[USER1].restyClient.R().SetBody(`{"albumTitle": "title", "albumDescription": "description"}`).SetResult(&respBody).Post(TEST_API + "/album/create")

	// Check that USER1 can access its album
	var album model.Album
	resp, err := clients[USER1].restyClient.R().SetQueryParam("albumId", respBody.AlbumId).SetResult(&album).Get(TEST_API + "/album/get")
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode())
	assert.Equal(t, "title", album.Title)
	assert.Equal(t, "description", album.Description)
	fmt.Println(album.AuthorId)

	// Check that USER2 cannot access USER1's album
	var album2 *model.Album
	resp2, err2 := clients[USER2].restyClient.R().SetQueryParam("albumId", respBody.AlbumId).SetResult(album2).Get(TEST_API + "/album/get")
	assert.Nil(t, err2)
	assert.Equal(t, 401, resp2.StatusCode())
	assert.Nil(t, album2)
}
