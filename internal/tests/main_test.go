package tests

import (
	"context"
	"data-storage-svc/internal/cli"
	"data-storage-svc/internal/database"
	"data-storage-svc/internal/deployment"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
)

type TestingUser struct {
	email       string
	password    string
	restyClient *resty.Client
}

const (
	UNAUTHENTICATED_CLIENT = "unauthenticated_client"
	USER1                  = "user1"
	USER2                  = "user2"
)

var clients = map[string]*TestingUser{
	UNAUTHENTICATED_CLIENT: {},
	USER1:                  {email: "user1@test.fr", password: "azerty1"},
	USER2:                  {email: "user2@test.fr", password: "azerty2"},
}

const (
	TEST_DB_NAME = "test_db"
	TEST_API     = "http://localhost:8080"
)

func setupForAllTests() {
	// Setup Test Db
	cli.DbName = TEST_DB_NAME
	slog.Debug("Using", "dbName", cli.DbName)
	deployment.StartMongoDB()

	// Clear the testing DB
	err := database.Mongo().Client().Database(TEST_DB_NAME).Drop(context.Background())
	if err != nil {
		slog.Error("couldn't clear the test database", "error", err)
		panic(err)
	}

	// Start Rest API
	go deployment.StartApi()
	time.Sleep(500000000)

	// Create predefined users for tests
	client := resty.New()
	for _, user := range clients {
		if len(user.email) > 0 && len(user.password) > 0 {
			// Create the user in the BDD
			resp, err := client.R().SetBody(fmt.Sprintf(`{"email": "%s", "password": "%s"}`, user.email, user.password)).Post(TEST_API + "/user/register")
			if err != nil {
				slog.Error("couldn't create test user", slog.Any("err", err))
				panic(err)
			} else if resp.StatusCode() != 201 {
				slog.Error("Couldn't create test user", "status_code", resp.StatusCode())
				panic(errors.New("couldn't create test user"))
			} else {
				slog.Debug("Created test user", "email", user.email, "password", user.password)
			}
			// Fetch JWT tokens for test users
			var respBody struct {
				Jwt string `json:"jwt"`
			}
			resp, err = client.R().SetBody(fmt.Sprintf(`{"email": "%s", "password": "%s"}`, user.email, user.password)).SetResult(&respBody).Post(TEST_API + "/user/jwt")
			if err != nil {
				slog.Error("couldn't fetch test user jwt", slog.Any("err", err))
				panic(err)
			} else if resp.StatusCode() != 200 {
				slog.Error("couldn't fetch test user jwt", "status_code", resp.StatusCode())
				panic(errors.New("couldn't fetch test user jwt"))
			}
			// Create ready-to-user resty clients for each test user
			newClient := resty.New()
			newClient.SetCookie(&http.Cookie{Name: "jwt", Value: respBody.Jwt})
			user.restyClient = newClient
		} else {
			// Unauthenticated user, no need to register or fetch jwt, just create a simple resty client
			user.restyClient = resty.New()
		}
	}
}

func TestMain(m *testing.M) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	setupForAllTests()
	slog.Debug("Running go tests ...")
	os.Exit(m.Run())
	slog.Debug("Cleaning after go tests")
}
