package main

import (
	"data-storage-svc/internal/deployment"
	"log/slog"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	slog.Info("Starting data-storage service...")
	deployment.StartMongoDB()
	// router := gin.Default()

}
