package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI      string
	DatabaseName  string
	ServerPort    string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, relying on system environment variables...")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI is required but not set")
	}

	databaseName := os.Getenv("MONGO_DB")
	if databaseName == "" {
		databaseName = "freelanceX_proposals" 
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = ":50051" 
	}

	return &Config{
		MongoURI:     mongoURI,
		DatabaseName: databaseName,
		ServerPort:   serverPort,
	}
}
