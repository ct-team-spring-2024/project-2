package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"oj/goforces/internal/middlewares"
	"oj/goforces/internal/models"
	"oj/goforces/internal/services"

	"github.com/sirupsen/logrus"
)

func CreateProblem(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var newProblem models.Problem
	if err := json.NewDecoder(r.Body).Decode(&newProblem); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	newProblem.OwnerId = userID
	createdProblem, err := services.CreateProblem(newProblem)
	logrus.Infof("CreatedProblem => %+v", createdProblem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdProblem)
}

func UpdateProblem(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var updatedProblem models.Problem
	if err := json.NewDecoder(r.Body).Decode(&updatedProblem); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	problem, err := services.UpdateProblem(userID, updatedProblem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(problem)
}

func GetMyProblems(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	problems := services.GetMyProblems(userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(problems)
	logrus.Info(problems)
}

func GetPublishedProblems(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pageStr := query.Get("page")
	pageSizeStr := query.Get("pageSize")
	page := 1
	pageSize := 10
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			page = p
		}
	}
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil {
			pageSize = ps
		}
	}
	problems := services.GetPublishedProblems(page, pageSize)
	logrus.Infof("FFFF => %+v", problems)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(problems)
}

func GetProblemByID(w http.ResponseWriter, r *http.Request) {

	segments := strings.Split(r.URL.Path, "/")
	if len(segments) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)

		return
	}
	idStr := segments[2]
	problemId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid problem ID", http.StatusBadRequest)
		return
	}

	problem, err := services.GetProblemById(problemId)
	if err != nil {
		http.Error(w, "Problem not found", http.StatusNotFound)

		return
	}
	// If not published, only allow viewing if the requester is the owner.
	if problem.Status != "Published" {
		userID, ok := middlewares.GetUserIDFromContext(r.Context())
		if !ok || problem.OwnerId != userID {
			http.Error(w, "Problem not published", http.StatusForbidden)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(problem)
}

func AdminGetAllProblems(w http.ResponseWriter, r *http.Request) {
	problems := services.GetAllProblems()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(problems)
}

func AdminUpdateProblemStatus(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ProblemId int    `json:"problemId"`
		NewStatus string `json:"newStatus"`
		Feedback  string `json:"feedback"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	logrus.Info("HERE")
	updatedProblem, err := services.UpdateProblemStatus(payload.ProblemId, payload.NewStatus, payload.Feedback)
	logrus.Info("HERE2")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedProblem)
}

func ProblemsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		CreateProblem(w, r)
	case http.MethodPut:
		UpdateProblem(w, r)
	case http.MethodGet:
		GetPublishedProblems(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
