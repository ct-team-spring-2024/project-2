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
		Username: "admin",
		Email:    "admin@email.com",
		Password: "admin",
		Role:     "admin",
	}
	u2 := models.User{
		Username: "user",
		Email:    "user@email.com",
		Password: "user",
		Role:     "user",
	}
	services.RegisterUser(u1)
	userId2, _ := services.RegisterUser(u2)
	problem1 := models.Problem{
		ProblemId:   0,
		OwnerId:     userId2,
		Title:       "problem 1 title",
		Statement:   "This is a simple problem",
		TimeLimit:   3000,
		MemoryLimit: 500,
		Inputs:      []string{"50 1 10", "50 1 600"},
		Outputs:     []string{"51", "51"},
		Status:      models.Published,
		Feedback:    "HA?",
		PublishDate: time.Now(),
	}
	services.CreateProblem(problem1)
	logrus.Info("WHAT?")
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

	//db.DB = db.NewXMockDB()
	db.DB = db.ConnectToDB()
	initSystem()
	// db.ConnectToDB()
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
