package api

import (
	"encoding/json"
	"net/http"
	"oj/goforces/internal"
	"oj/goforces/internal/controllers"
	"oj/goforces/internal/db"
	"oj/goforces/internal/middlewares"
	"oj/goforces/internal/submission"

	"github.com/sirupsen/logrus"
)

var DB db.Database

func helloWorld(w http.ResponseWriter, r *http.Request) {
	response := "gooz"
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

// TODO: should extract the token from request to identify the current user.
// TODO: should return the value in the http response
func GetUserSubmission(w http.ResponseWriter, r *http.Request) {
	u := internal.NewUser(1, "u1", "p1")
	submissions := submission.GetUserSubmission(DB, *u)
	logrus.Infof("All user submissions => %+v", submissions)

	w.Header().Set("Content-Type", "application/json")
	if len(submissions) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "No submissions found"}`))
		return
	}
	if err := json.NewEncoder(w).Encode(submissions); err != nil {
		logrus.Errorf("Failed to encode submissions to JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AddSubmission(w http.ResponseWriter, r *http.Request) {
	s := internal.NewSubmission(1, 100)
	err := submission.AddSubmission(DB, *s)
	if err != nil {
		logrus.Fatalf("Err => %v", err)
	}
	logrus.Info("Submit added successfully")
}

func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", helloWorld)
	mux.HandleFunc("/user-submission", GetUserSubmission)
	mux.HandleFunc("/add-submission", AddSubmission)

	mux.HandleFunc("/register", controllers.Register)
	mux.HandleFunc("/login", controllers.Login)

	mux.Handle("/profile", middlewares.AuthMiddleware(http.HandlerFunc(controllers.GetProfile)))
	mux.Handle("/profile/update", middlewares.AuthMiddleware(http.HandlerFunc(controllers.UpdateProfile)))
	return mux
}
