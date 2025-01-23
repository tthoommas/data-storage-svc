package main

import (
	"context"
	"data-storage-svc/internal"
	"data-storage-svc/internal/deployment"
	"log"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:        "album",
		Usage:       "A media album API",
		Description: "A media album API to store and share medias",
		Commands: []*cli.Command{
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "Run the photo album API",
				Action: func(ctx context.Context, c *cli.Command) error {
					if internal.DEBUG {
						slog.SetLogLoggerLevel(slog.LevelDebug)
					} else {
						slog.SetLogLoggerLevel(slog.LevelError)
					}
					deployment.StartApi()
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "mongo-url",
						Usage:       "The mongo connection string",
						Value:       "mongodb://localhost:27017/",
						Destination: &internal.MONGO_URL,
					},
					&cli.StringFlag{
						Name:        "db-name",
						Usage:       "The mongo database name to use",
						Value:       "db",
						Destination: &internal.DB_NAME,
					},
					&cli.StringFlag{
						Name:        "api-ip",
						Usage:       "The IP to expose the API against",
						Value:       "127.0.0.1",
						Destination: &internal.API_IP,
					},
					&cli.IntFlag{
						Name:        "api-port",
						Usage:       "The port to expose the API against",
						Value:       8080,
						Destination: &internal.API_PORT,
					},
					&cli.BoolFlag{
						Name:        "debug",
						Aliases:     []string{"d"},
						Value:       false,
						Destination: &internal.DEBUG,
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
