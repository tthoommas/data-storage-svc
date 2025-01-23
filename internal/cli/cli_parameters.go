package cli

import (
	"flag"
	"log/slog"
)

var DbName = "db"
var ApiIp = "0.0.0.0"
var ApiPort = 8080
var MongoAutoStart = true
var MongoConnectionString = "mongodb://localhost:27017"

func LoadCliParameters() {
	slog.Debug("Loading cli parameters")
	dbName := flag.String("db", "db", "The Mongo DB name to use to store API data")
	apiIp := flag.String("ip", "0.0.0.0", "The IP to bind the HTTP API on")
	apiPort := flag.Int("port", 8080, "The HTTP API port to use")
	mongoAutoStart := flag.Bool("mongo", true, "Automatically start mongo DB docker image if needed")
	mongoConnectionString := flag.String("mongo-url", "mongodb://localhost:27017", "The mongo connection string.")
	flag.Parse()

	DbName = *dbName
	ApiIp = *apiIp
	ApiPort = *apiPort
	MongoAutoStart = *mongoAutoStart
	MongoConnectionString = *mongoConnectionString
}
