package main_test

// import (
//	"encoding/json"
//	"net/http"
//	"net/http/httptest"
//	"testing"

//	"github.com/sirupsen/logrus"

//	"oj/goforces/api"
//	"oj/goforces/internal/db"
//	"oj/goforces/internal/models"
// )

// func TestAddCorrectSubmit(t *testing.T) {
//	logger := logrus.New()
//	logger.SetLevel(logrus.DebugLevel)
//	logger.SetFormatter(&logrus.TextFormatter{})

//	logger.Debug("This is a debug message")

//	port := ":8080"

//	// Setup DB
//	db.DB = db.NewXMockDB()

//	router := api.SetupRoutes()

//	// Start the HTTP server.
//	logrus.Infof("Starting API server on port %s...\n", port)
//	server := httptest.NewServer(router)

//	http.Get(server.URL + "/add-submission")
//	http.Get(server.URL + "/add-submission")

//	var submissions []models.Submission
//	resp, _ := http.Get(server.URL + "/user-submission")
//	if err := json.NewDecoder(resp.Body).Decode(&submissions); err != nil {
//		// Handle decoding error
//	}
//	logrus.Infof("user submissionssss => %+v", submissions)
//	// TODO: evaluating the code, and checking if the status is OK

//	server.Close()
// }
