package controllers

import (
	"encoding/json"
	"net/http"

	"oj/goforces/internal/models"
	"oj/goforces/internal/services"

	"github.com/sirupsen/logrus"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var newUser models.User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	userId, err := services.RegisterUser(newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	token, err := services.GenerateToken(newUser)
	if err != nil {
		logrus.Error("Error generating token for user")
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"userId": userId, "token": token})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	token, err := services.AuthenticateUserWithUsername(credentials.Username, credentials.Password)

	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
