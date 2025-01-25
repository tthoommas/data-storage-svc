package deployment

import (
	"context"
	"data-storage-svc/internal/utils"
	"errors"
	"io"
	"log/slog"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
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

	containers, err := apiClient.ContainerList(context.Background(), container.ListOptions{All: true, Filters: filters.NewArgs(filters.KeyValuePair{Key: "ancestor", Value: "mongo:4.4.1"})})
	if err != nil {
		slog.Error("Couldn't list docker containers")
		panic(err)
	}

	var mongoContainerID string
	mongoContainerRunning := false

	if len(containers) == 0 {
		slog.Debug("No existing mongo container found, creating one")

		// Pull the MongoDB image
		out, err := apiClient.ImagePull(context.Background(), "mongo:4.4.1", image.PullOptions{})
		if err != nil {
			slog.Error("couldn't pull the mongo image")
			panic(err)
		}
		io.Copy(os.Stdout, out)
		defer out.Close()

		containerConfig := &container.Config{
			Image:    "mongo:4.4.1",
			Hostname: "mongo",
		}
		dataDir, err := utils.GetDataDir("mongo")
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
