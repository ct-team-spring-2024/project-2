package main

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"oj/goforces/api"
	"oj/goforces/internal/db"
)


func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{})

	logger.Debug("This is a debug message")

	port := ":8080"

	router := api.SetupRoutes()
	// Setup DB
	api.DB = db.NewMockDB()

	// Start the HTTP server.
	logrus.Infof("Starting API server on port %s...\n", port)
	logrus.Fatal(http.ListenAndServe(port, router))
}
