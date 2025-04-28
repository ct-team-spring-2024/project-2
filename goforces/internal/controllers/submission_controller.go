package controllers

import (
	"encoding/json"
	"net/http"

	"oj/goforces/internal/middlewares"
	"oj/goforces/internal/services"

	"github.com/sirupsen/logrus"
)

func CreateSubmission(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	isSync := queryParams.Get("sync") == "true"

	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var payload struct {
		ProblemId int    `json:"problemId"`
		Code      string `json:"code"`
		Language  string `json:"language"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	submissionId, err := services.CreateSubmission(userID, payload.ProblemId, payload.Code, payload.Language)
	logrus.Infof("submissionId => %+v", submissionId)
	if err != nil {
		logrus.Info("#0")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logrus.Info("#1")
	// Send the code to the judge service
	problem, err := services.GetProblemById(payload.ProblemId)
	logrus.Infof("probID %d", payload.ProblemId)
	logrus.Infof("ERRRR %+v", err)
	if err != nil {
		logrus.Info("#2")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if isSync {
		services.EvalCode(submissionId, problem)
		logrus.Infof("sub Id %d", submissionId)
	} else {
		go services.EvalCode(submissionId, problem)
	}
	logrus.Infof("GGGG")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"submissionId": submissionId,
		"status":       "Processing",
	})
}

func GetMySubmissions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	submissions, err := services.GetSubmissionsByUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submissions)
}
