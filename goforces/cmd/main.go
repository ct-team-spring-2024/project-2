package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"oj/goforces/api"
	"oj/goforces/internal/config"
	"oj/goforces/internal/db"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{})

	logger.Debug("This is a debug message")

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Error loading configuration: %v", err)
	}
	// if err := db.Connect(cfg.DatabaseURL); err != nil {
	// 	Logger.Fatalf("Error connecting to the database: %v", err)
	// }

	// if err := db.Migrate(); err != nil {
	// 	logger.Fatalf("Error during database migration: %v", err)
	// }

	router := api.SetupRoutes()

	port := cfg.Port

	api.DB = db.NewMockDB()

	serverAddress := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:         serverAddress,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	// TODO: GoForces status
	logger.Printf("Server is running on port %d", port)
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatalf("Server error: %v", err)
	}

	// logger.Infof("Starting API server on port %s...\n", port)
	// logger.Fatal(http.ListenAndServe(strconv.Itoa(port), router))
}
