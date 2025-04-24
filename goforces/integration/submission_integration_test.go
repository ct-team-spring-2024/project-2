package main_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"oj/goforces/api"
	"oj/goforces/internal/db"
)

func TestUserSubmissions(t *testing.T) {
	handler := api.SetupRoutes()
	db.DB = db.NewXMockDB()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	token := loginTestUser(t, ts)

	req, err := http.NewRequest("GET", ts.URL+"/submissions", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if len(body) == 0 {
		t.Error("Expected non-empty response body")
	}
}

func TestUserSubmissionStats(t *testing.T) {
	handler := api.SetupRoutes()
	db.DB = db.NewXMockDB()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	token := loginTestUser(t, ts)

	req, err := http.NewRequest("GET", ts.URL+"/profile", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if len(body) == 0 {
		t.Error("Expected non-empty response body")
	}
}

func TestEmptySubmission(t *testing.T) {
	handler := api.SetupRoutes()
	db.DB = db.NewXMockDB()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	token := loginTestUser(t, ts)

	payload := map[string]interface{}{
		"problemId": 1,
		"code":      "",
		"language":  "golang",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", ts.URL+"/submit", bytes.NewReader(payloadBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status BadRequest for empty submission; got %v", resp.Status)
	}
}

// TODO: We are not currenlty handing empty submissions error case

func loginTestUser(t *testing.T, ts *httptest.Server) string {
	regPayload := map[string]string{
		"username": "subuser",
		"email":    "subuser@example.com",
		"password": "password",
		"role":     "user",
	}
	regBytes, err := json.Marshal(regPayload)
	if err != nil {
		t.Fatalf("Failed to marshal registration payload: %v", err)
	}
	_, err = http.Post(ts.URL+"/register", "application/json", bytes.NewReader(regBytes))
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	loginPayload := map[string]string{
		"email":    "subuser@example.com",
		"password": "password",
	}
	loginBytes, err := json.Marshal(loginPayload)
	if err != nil {
		t.Fatalf("Failed to marshal login payload: %v", err)
	}
	resp, err := http.Post(ts.URL+"/login", "application/json", bytes.NewReader(loginBytes))
	if err != nil {
		t.Fatalf("Failed to login test user: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read login response: %v", err)
	}
	var loginResult map[string]string
	if err := json.Unmarshal(body, &loginResult); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}

	token := loginResult["token"]
	if token == "" {
		t.Fatal("No token received from login")
	}
	return token
}
