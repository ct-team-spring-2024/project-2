package main_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oj/goforces/api"
	"oj/goforces/internal/db"
)

func TestUserRegistration(t *testing.T) {
	handler := api.SetupRoutes()
	db.DB = db.NewXMockDB()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	payload := map[string]string{
		"username": "testuser",
		"email":    "testuser@example.com",
		"password": "password123",
		"role":     "user",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}
	resp, err := http.Post(ts.URL+"/register", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status OK, got %v. Response: %s", resp.Status, string(body))
	}
}

func TestUserLogin(t *testing.T) {
	handler := api.SetupRoutes()
	db.DB = db.NewXMockDB()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// TestUserRegistration(t)
	time.Sleep(100 * time.Millisecond)

	payload := map[string]string{
		"email":    "testuser@example.com",
		"password": "password123",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}
	resp, err := http.Post(ts.URL+"/login", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status OK, got %v. Response: %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result["token"] == "" {
		t.Error("Expected token in response, got none")
	}
}

func TestUserProfile(t *testing.T) {
	handler := api.SetupRoutes()
	db.DB = db.NewXMockDB()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// TestUserRegistration(t)
	time.Sleep(100 * time.Millisecond)

	payload := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(ts.URL+"/login", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)
	token := result["token"]

	req, _ := http.NewRequest("GET", ts.URL+"/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status OK, got %v. Response: %s", resp.Status, string(body))
	}
}

func TestUserProfileUpdate(t *testing.T) {
	handler := api.SetupRoutes()
	db.DB = db.NewXMockDB()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// TestUserRegistration(t)
	time.Sleep(100 * time.Millisecond)

	payload := map[string]string{
		"email":    "testuser@example.com",
		"password": "password123",
	}
	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(ts.URL+"/login", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)
	token := result["token"]

	updatePayload := map[string]string{
		"username": "updateduser",
		"password": "newpassword123",
	}
	updatePayloadBytes, err := json.Marshal(updatePayload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	req, _ := http.NewRequest("POST", ts.URL+"/profile/update", bytes.NewReader(updatePayloadBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status OK, got %v. Response: %s", resp.Status, string(body))
	}
}
