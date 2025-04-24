package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"oj/goforces/internal/middlewares"
	"oj/goforces/internal/models"
	"oj/goforces/internal/services"

)

func GetProfile(w http.ResponseWriter, r *http.Request) {

	username := r.PathValue("username")
	logrus.Infof("username => %s", username)

	user, _ := services.GetUserByUsername(username)
	stats := services.GetSubmissionStats(user)

	user.Password = ""

	response := map[string]interface{}{
		"profile":         user,
		"submissionStats": stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	var updatedData struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updatedData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user, err := services.UpdateUserProfile(userID,
		models.User{
			UserId:   userID,
			Username: updatedData.Username,
			Password: updatedData.Password,
			Role:     updatedData.Role,
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
