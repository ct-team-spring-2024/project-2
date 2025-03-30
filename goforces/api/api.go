package api

import (
	"net/http"
	"oj/goforces/internal"
	"oj/goforces/internal/db"
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
	result := submission.GetUserSubmission(DB, *u)
	logrus.Infof("result => %v", result)
}

func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", helloWorld)
	mux.HandleFunc("/user-submission", GetUserSubmission)
	return mux
}
