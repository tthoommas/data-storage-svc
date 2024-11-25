package cli

import (
	"flag"
	"log/slog"
)

var DbName = "db"

func LoadCliParameters() {
	slog.Debug("Loading cli parameters")
	dbName := flag.String("db", "db", "The Mongo DB name to use to store API data")
	flag.Parse()
	DbName = *dbName
	slog.Debug("CLI param", "DB name", DbName)
}
