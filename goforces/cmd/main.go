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
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})

	logrus.Debug("This is a debug message")

	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("Error loading configuration: %v", err)
	}

	router := api.SetupRoutes()

	port := cfg.Port

	db.DB = db.NewXMockDB()

	serverAddress := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:         serverAddress,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	// TODO: GoForces status
	logrus.Printf("Server is running on port %d", port)
	if err := srv.ListenAndServe(); err != nil {
		logrus.Fatalf("Server error: %v", err)
	}
}
