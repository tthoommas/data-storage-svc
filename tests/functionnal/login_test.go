package functionnal

import (
	"data-storage-svc/internal/cli"
	"data-storage-svc/internal/deployment"
	"testing"

	"github.com/go-resty/resty/v2"
)

func TestLogin(t *testing.T) {
	cli.DbName = "test"
	go deployment.StartApi()
	client := resty.New()

	resp, err := client.R().
		EnableTrace().
		Get("https://httpbin.org/get")

}
