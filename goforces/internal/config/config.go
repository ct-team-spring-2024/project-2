package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        int
	DatabaseURL string
}

func LoadConfig() (*Config, error) {

	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "mockdb"
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
	}, nil
}
