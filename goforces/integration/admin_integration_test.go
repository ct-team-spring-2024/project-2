package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oj/goforces/api"
	"oj/goforces/internal/db"
)

func adminRegisterAndLogin(ts *httptest.Server) (string, int) {
	regPayload := map[string]string{
		"username": "adminuser",
		"email":    "admin@example.com",
		"password": "adminpass",
		"role":     "admin",
	}
	regBytes, _ := json.Marshal(regPayload)
	http.Post(ts.URL+"/register", "application/json", bytes.NewReader(regBytes))

	loginPayload := map[string]string{
		"email":    "admin@example.com",
		"password": "adminpass",
	}
	loginBytes, _ := json.Marshal(loginPayload)
	resp, _ := http.Post(ts.URL+"/login", "application/json", bytes.NewReader(loginBytes))
	body, _ := ioutil.ReadAll(resp.Body)
	var loginResult map[string]string
	json.Unmarshal(body, &loginResult)
	token := loginResult["token"]
	return token, 2
}

func userRegisterAndLogin(ts *httptest.Server, username, email, password string) (string, int) {
	regPayload := map[string]string{
		"username": username,
		"email":    email,
		"password": password,
		"role":     "user",
	}
	regBytes, _ := json.Marshal(regPayload)
	http.Post(ts.URL+"/register", "application/json", bytes.NewReader(regBytes))
	loginPayload := map[string]string{
		"email":    email,
		"password": password,
	}
	loginBytes, _ := json.Marshal(loginPayload)
	resp, _ := http.Post(ts.URL+"/login", "application/json", bytes.NewReader(loginBytes))
	body, _ := ioutil.ReadAll(resp.Body)
	var loginResult map[string]interface{}
	json.Unmarshal(body, &loginResult)
	token := loginResult["token"].(string)
	return token, 3
}

func TestAdminChangeUserRole(t *testing.T) {
	handler := api.SetupRoutes()
	db.DB = db.NewXMockDB()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	adminToken, _ := adminRegisterAndLogin(ts)
	_, targetUserId := userRegisterAndLogin(ts, "normalUser", "normal@example.com", "password")

	fmt.Println("Running TestAdminChangeUserRole...")
	payload := map[string]interface{}{
		"userId": targetUserId,
		"role":   "admin",
	}
	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", ts.URL+"/admin/user/role", bytes.NewReader(payloadBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Change User Role Response:", string(body))
}

func TestAdminChangeOwnRole(t *testing.T) {
	handler := api.SetupRoutes()
	db.DB = db.NewXMockDB()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	adminToken, adminUserId := adminRegisterAndLogin(ts)

	fmt.Println("Running TestAdminChangeOwnRole...")
	payload := map[string]interface{}{
		"userId": adminUserId,
		"role":   "user",
	}
	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", ts.URL+"/admin/user/role", bytes.NewReader(payloadBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Change Own Role (should fail) Response:", string(body))
}

func TestAdminPublishProblemDraft(t *testing.T) {
	handler := api.SetupRoutes()
	db.DB = db.NewXMockDB()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	adminToken, _ := adminRegisterAndLogin(ts)

	fmt.Println("Running TestAdminPublishProblemDraft...")
	// Create a problem draft as a normal user.
	regPayload := map[string]string{
		"username": "probowner",
		"email":    "probowner@example.com",
		"password": "password",
		"role":     "user",
	}
	regBytes, _ := json.Marshal(regPayload)
	http.Post(ts.URL+"/register", "application/json", bytes.NewReader(regBytes))
	loginPayload := map[string]string{
		"email":    "probowner@example.com",
		"password": "password",
	}
	loginBytes, _ := json.Marshal(loginPayload)
	resp, _ := http.Post(ts.URL+"/login", "application/json", bytes.NewReader(loginBytes))
	body, _ := ioutil.ReadAll(resp.Body)
	var loginResult map[string]interface{}
	json.Unmarshal(body, &loginResult)
	ownerToken := loginResult["token"].(string)

	// Create problem draft.
	probPayload := map[string]interface{}{
		"title":       "Sample Problem Draft",
		"statement":   "Solve x+y.",
		"timeLimit":   1,
		"memoryLimit": 128,
		"input":       "two ints",
		"output":      "their sum",
	}
	probBytes, _ := json.Marshal(probPayload)
	reqProb, _ := http.NewRequest("POST", ts.URL+"/problems", bytes.NewReader(probBytes))
	reqProb.Header.Set("Authorization", "Bearer "+ownerToken)
	reqProb.Header.Set("Content-Type", "application/json")
	respProb, _ := http.DefaultClient.Do(reqProb)
	bodyProb, _ := ioutil.ReadAll(respProb.Body)
	var probResult map[string]interface{}
	json.Unmarshal(bodyProb, &probResult)
	problemId := int(probResult["problemId"].(float64))
	fmt.Printf("Created Problem ID: %d\n", problemId)

	// Admin publishes the problem.
	statusPayload := map[string]interface{}{
		"problemId": problemId,
		"newStatus": "published",
		"feedback":  "",
	}
	statusBytes, _ := json.Marshal(statusPayload)
	reqStatus, _ := http.NewRequest("POST", ts.URL+"/admin/problems/status", bytes.NewReader(statusBytes))
	reqStatus.Header.Set("Authorization", "Bearer "+adminToken)
	reqStatus.Header.Set("Content-Type", "application/json")
	respStatus, _ := http.DefaultClient.Do(reqStatus)
	bodyStatus, _ := ioutil.ReadAll(respStatus.Body)
	fmt.Println("Publish Response:", string(bodyStatus))
}

func TestAdminRejectProblemDraft(t *testing.T) {
	handler := api.SetupRoutes()
	db.DB = db.NewXMockDB()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	adminToken, _ := adminRegisterAndLogin(ts)

	fmt.Println("Running TestAdminRejectProblemDraft...")
	regPayload := map[string]string{
		"username": "draftowner",
		"email":    "draftowner@example.com",
		"password": "password",
		"role":     "user",
	}
	regBytes, _ := json.Marshal(regPayload)
	http.Post(ts.URL+"/register", "application/json", bytes.NewReader(regBytes))
	loginPayload := map[string]string{
		"email":    "draftowner@example.com",
		"password": "password",
	}
	loginBytes, _ := json.Marshal(loginPayload)
	resp, _ := http.Post(ts.URL+"/login", "application/json", bytes.NewReader(loginBytes))
	body, _ := ioutil.ReadAll(resp.Body)
	var loginResult map[string]interface{}
	json.Unmarshal(body, &loginResult)
	ownerToken := loginResult["token"].(string)

	probPayload := map[string]interface{}{
		"title":       "Problem Draft to be Rejected",
		"statement":   "Do something.",
		"timeLimit":   2,
		"memoryLimit": 256,
		"input":       "data",
		"output":      "result",
	}
	probBytes, _ := json.Marshal(probPayload)
	reqProb, _ := http.NewRequest("POST", ts.URL+"/problems", bytes.NewReader(probBytes))
	reqProb.Header.Set("Authorization", "Bearer "+ownerToken)
	reqProb.Header.Set("Content-Type", "application/json")
	respProb, _ := http.DefaultClient.Do(reqProb)
	bodyProb, _ := ioutil.ReadAll(respProb.Body)
	var probResult map[string]interface{}
	json.Unmarshal(bodyProb, &probResult)
	problemId := int(probResult["problemId"].(float64))
	fmt.Printf("Created Draft Problem ID: %d\n", problemId)

	statusPayload := map[string]interface{}{
		"problemId": problemId,
		"newStatus": "rejected",
		"feedback":  "Insufficient problem statement details.",
	}
	statusBytes, _ := json.Marshal(statusPayload)
	reqStatus, _ := http.NewRequest("POST", ts.URL+"/admin/problems/status", bytes.NewReader(statusBytes))
	reqStatus.Header.Set("Authorization", "Bearer "+adminToken)
	reqStatus.Header.Set("Content-Type", "application/json")
	respStatus, _ := http.DefaultClient.Do(reqStatus)
	bodyStatus, _ := ioutil.ReadAll(respStatus.Body)
	fmt.Println("Reject Response:", string(bodyStatus))
}

func TestAdminChangePublishedToDraft(t *testing.T) {
	handler := api.SetupRoutes()
	db.DB = db.NewXMockDB()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	adminToken, _ := adminRegisterAndLogin(ts)

	fmt.Println("Running TestAdminChangePublishedToDraft...")

	// First publish a problem
	TestAdminPublishProblemDraft(t)
	time.Sleep(100 * time.Millisecond)

	// Then change it back to draft
	problemId := 1 // Assuming the first problem has ID 1
	statusPayload := map[string]interface{}{
		"problemId": problemId,
		"newStatus": "draft",
		"feedback":  "Reverting due to formatting issues.",
	}
	statusBytes, _ := json.Marshal(statusPayload)
	reqStatus, _ := http.NewRequest("POST", ts.URL+"/admin/problems/status", bytes.NewReader(statusBytes))
	reqStatus.Header.Set("Authorization", "Bearer "+adminToken)
	reqStatus.Header.Set("Content-Type", "application/json")
	respStatus, _ := http.DefaultClient.Do(reqStatus)
	bodyStatus, _ := ioutil.ReadAll(respStatus.Body)
	fmt.Println("Change to Draft Response:", string(bodyStatus))
}
