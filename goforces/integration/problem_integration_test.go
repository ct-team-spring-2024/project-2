package main_test

// import (
//	"bytes"
//	"encoding/json"
//	"io/ioutil"
//	"net/http"
//	"net/http/httptest"
//	"testing"

//	"oj/goforces/api"
//	"oj/goforces/internal/db"
// )

// func setupProblemTestServer(t *testing.T) (*httptest.Server, string) {
//	handler := api.SetupRoutes()
//	db.DB = db.NewXMockDB()
//	ts := httptest.NewServer(handler)

//	regPayload := map[string]string{
//		"username": "probuser",
//		"email":    "probuser@example.com",
//		"password": "password",
//		"role":     "user",
//	}
//	regBytes, _ := json.Marshal(regPayload)
//	http.Post(ts.URL+"/register", "application/json", bytes.NewReader(regBytes))

//	loginPayload := map[string]string{
//		"email":    "probuser@example.com",
//		"password": "password",
//	}
//	loginBytes, _ := json.Marshal(loginPayload)
//	resp, err := http.Post(ts.URL+"/login", "application/json", bytes.NewReader(loginBytes))
//	if err != nil {
//		t.Fatalf("Failed to login: %v", err)
//	}

//	body, _ := ioutil.ReadAll(resp.Body)
//	var loginResult map[string]string
//	json.Unmarshal(body, &loginResult)
//	return ts, loginResult["token"]
// }

// func TestProblemsPagePagination(t *testing.T) {
//	ts, token := setupProblemTestServer(t)
//	defer ts.Close()

//	url := ts.URL + "/problems?page=1&pageSize=10"
//	req, err := http.NewRequest("GET", url, nil)
//	if err != nil {
//		t.Fatalf("Failed to create request: %v", err)
//	}
//	req.Header.Set("Authorization", "Bearer "+token)

//	resp, err := http.DefaultClient.Do(req)
//	if err != nil {
//		t.Fatalf("Failed to get first page: %v", err)
//	}
//	defer resp.Body.Close()

//	if resp.StatusCode != http.StatusOK {
//		t.Errorf("Expected status OK; got %v", resp.Status)
//	}

//	// Test second page
//	url2 := ts.URL + "/problems?page=2&pageSize=10"
//	req2, err := http.NewRequest("GET", url2, nil)
//	if err != nil {
//		t.Fatalf("Failed to create second request: %v", err)
//	}
//	req2.Header.Set("Authorization", "Bearer "+token)

//	resp2, err := http.DefaultClient.Do(req2)
//	if err != nil {
//		t.Fatalf("Failed to get second page: %v", err)
//	}
//	defer resp2.Body.Close()

//	if resp2.StatusCode != http.StatusOK {
//		t.Errorf("Expected status OK; got %v", resp2.Status)
//	}
// }

// func TestMyProblems(t *testing.T) {
//	ts, token := setupProblemTestServer(t)
//	defer ts.Close()

//	req, err := http.NewRequest("GET", ts.URL+"/problems/mine", nil)
//	if err != nil {
//		t.Fatalf("Failed to create request: %v", err)
//	}
//	req.Header.Set("Authorization", "Bearer "+token)

//	resp, err := http.DefaultClient.Do(req)
//	if err != nil {
//		t.Fatalf("Failed to get my problems: %v", err)
//	}
//	defer resp.Body.Close()

//	if resp.StatusCode != http.StatusOK {
//		t.Errorf("Expected status OK; got %v", resp.Status)
//	}
// }
