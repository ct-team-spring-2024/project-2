package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"oj/goforces/api"
	"oj/goforces/internal/config"
	"oj/goforces/internal/db"
	"oj/goforces/internal/models"
	"oj/goforces/internal/services"
)

func initSystem() {
	u1 := models.User{
		UserId:   0,
		Username: "testuser",
		Email:    "testuser@email.com",
		Password: "password123",
		Role:     "user",
	}
	services.RegisterUser(u1)
	problem1 := models.Problem{
		ProblemId:   1,
		OwnerId:     u1.UserId,
		Title:       "problem 1 title",
		Statement:   "This is a simple problem",
		TimeLimit:   1000,
		MemoryLimit: 2000,
		Input:       "KIR",
		Output:      "KHAR",
		Status:      "published",
		Feedback:    "HA?",
		PublishDate: time.Now(),
	}

	services.CreateProblem(problem1)
}

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
	initSystem()

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
