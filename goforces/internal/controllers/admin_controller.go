package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"oj/goforces/internal/middlewares"
	"oj/goforces/internal/models"
	"oj/goforces/internal/services"
)

func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.URL.Query().Get("userId")
	if userIdStr == "" {
		http.Error(w, "userId parameter required", http.StatusBadRequest)
		return
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, "invalid userId", http.StatusBadRequest)
		return
	}
	user, err := services.GetUserByID(userId)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	stats := services.GetSubmissionStats(user)

	user.Password = ""

	response := map[string]interface{}{
		"profile":         user,
		"submissionStats": stats,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		UserId int    `json:"userId"`
		Role   string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	adminID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User not in context", http.StatusUnauthorized)
		return
	}

	if adminID == payload.UserId {
		http.Error(w, "Cannot change your own role", http.StatusForbidden)
		return
	}

	user, err := services.GetUserByID(payload.UserId)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if payload.Role != "admin" && payload.Role != "user" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	updatedUser, err := services.UpdateUserProfile(user.UserId, models.User{
		UserId:   user.UserId,
		Username: user.Username,
		Email:    user.Email,
		Password: user.Password,
		Role:     payload.Role,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	updatedUser.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}
