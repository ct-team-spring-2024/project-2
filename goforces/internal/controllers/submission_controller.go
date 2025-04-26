package controllers

import (
	"encoding/json"
	"net/http"

	"oj/goforces/internal/middlewares"
	"oj/goforces/internal/services"
)

func CreateSubmission(w http.ResponseWriter, r *http.Request) {
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

	submission, err := services.CreateSubmission(userID, payload.ProblemId, payload.Code, payload.Language)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Send the code to the judge service
	problem, err := services.GetProblemById(payload.ProblemId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	go services.EvalCode(submission, problem)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submission)
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
