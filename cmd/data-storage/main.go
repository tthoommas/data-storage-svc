package main

import (
	"data-storage-svc/internal/cli"
	"data-storage-svc/internal/deployment"
	"log/slog"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	cli.LoadCliParameters()
	if cli.MongoAutoStart {
		deployment.StartMongoDB()
	}
	deployment.StartApi()
}
