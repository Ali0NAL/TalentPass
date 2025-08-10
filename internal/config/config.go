package config

import (
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dbURL := os.Getenv("DATABASE_URL")
	// docker-compose ile default DB:
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/talentpass?sslmode=disable"
	}
	return Config{
		Port:        port,
		DatabaseURL: dbURL,
	}
}
