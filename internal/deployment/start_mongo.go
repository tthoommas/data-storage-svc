package deployment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

func StartMongoDB() error {
	slog.Debug("Starting Mongo DB")
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	slog.Debug("Docker client created v" + apiClient.ClientVersion())
	defer apiClient.Close()

	containers, err := apiClient.ContainerList(context.Background(), container.ListOptions{All: true, Filters: filters.NewArgs(filters.KeyValuePair{Key: "ancestor", Value: "mongo"})})
	if err != nil {
		slog.Error("Couldn't list docker containers")
		panic(err)
	}

	var mongoContainerID string
	mongoContainerRunning := false

	if len(containers) == 0 {
		slog.Debug("No existing mongo container found, creating one")
		containerConfig := &container.Config{
			Image:    "mongo",
			Hostname: "mongo",
		}
		dataDir, err := getDataDir()
		if err != nil {
			slog.Error("couldn't find the /data directory ")
			panic(err)
		} else {
			slog.Debug("Found data dir for mongo data", "path", dataDir)
		}
		hostConfig := &container.HostConfig{
			NetworkMode: "host",
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: dataDir,
					Target: "/data/db",
				},
			},
		}
		response, err := apiClient.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, nil, "mongo")
		if err != nil {
			slog.Error("couldn't create the mongo DB container")
			panic(err)
		} else {
			slog.Debug("Created mongo DB container", "id", response.ID)
			mongoContainerID = response.ID
		}
	} else if len(containers) > 1 {
		slog.Error("Found several mongo containers")
		return errors.New("found multiple existing mongo containers")
	} else {
		// Found exactly one mongo container
		mongoContainerID = containers[0].ID
		mongoContainerRunning = containers[0].State == "running"
		slog.Debug("Found existing mongo DB container", "id", mongoContainerID, "running", mongoContainerRunning)
	}

	if !mongoContainerRunning {
		// Start the non running container
		slog.Debug("Starting non running mongo DB container", "id", mongoContainerID)
		if err := apiClient.ContainerStart(context.Background(), mongoContainerID, container.StartOptions{}); err != nil {
			slog.Error("couldn't start mongo db container", "id", mongoContainerID)
			panic(err)
		}
		slog.Info("Mongo DB container started", "containerId", mongoContainerID)
	} else {
		slog.Info("Mongo DB container already running", "containerId", mongoContainerID)
	}

	return nil
}

func getDataDir() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", fmt.Errorf("unable to get current file info")
	}
	currentFilePath := filepath.Dir(filename)
	parts := strings.Split(currentFilePath, string(filepath.Separator))

	// Find the index of "internal"
	index := -1
	for i, part := range parts {
		if part == "internal" {
			index = i
			break
		}
	}

	if index == -1 {
		return "", errors.New("The segment 'internal' was not found in the path")
	}

	// Replace "internal" with "data/mongo"
	parts[index] = "data/mongo"
	parts = append([]string{"/"}, parts...)

	dataDirPath := filepath.Join(parts[:index+2]...)
	exists, err := pathExists(dataDirPath)
	if err != nil || !exists {
		return "", errors.New("data dir path do not exists")
	}

	return dataDirPath, nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil // path exists
	}
	if os.IsNotExist(err) {
		return false, nil // path does not exist
	}
	return false, err // some other error
}
