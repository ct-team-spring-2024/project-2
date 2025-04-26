package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"oj/goforces/internal/db"
	"oj/goforces/internal/models"
)

// We need to do a for loop until the submission reaches a terminal state.
func EvalCode(s models.Submission, p models.Problem) {
	retryCnt := 0
	evalCodeUrl := "http://localhost:8082/eval-code"
	for {
		payload := map[string]interface{}{
			"code":        s.Code,
			"inputs":      p.Inputs,
			"timelimit":   p.TimeLimit,
			"memorylimit": p.MemoryLimit,
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			logrus.Errorf("Failed to marshal payload: %v", err)
			retryCnt++
			time.Sleep(time.Second * time.Duration(retryCnt)) // Exponential backoff
			continue
		}

		// Create the HTTP POST request
		req, err := http.NewRequest("POST", evalCodeUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			logrus.Errorf("Failed to create HTTP request: %v", err)
			retryCnt++
			time.Sleep(time.Second * time.Duration(retryCnt)) // Exponential backoff
			continue
		}

		// Set the required headers
		req.Header.Set("Content-Type", "application/json")

		// Send the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			logrus.Errorf("Failed to send HTTP request: %v", err)
			retryCnt++
			time.Sleep(time.Second * time.Duration(retryCnt)) // Exponential backoff
			continue
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logrus.Errorf("Failed to read response body: %v", err)
			retryCnt++
			time.Sleep(time.Second * time.Duration(retryCnt)) // Exponential backoff
			continue
		}

		// Log the response for debugging
		logrus.Infof("Response from eval-code endpoint: %s", string(body))

		// Parse the response
		var results []struct {
			Status string `json:"Status"`
			Output string `json:"Output"`
		}

		err = json.Unmarshal(body, &results)
		if err != nil {
			logrus.Errorf("Failed to unmarshal response: %v", err)
			retryCnt++
			time.Sleep(time.Second * time.Duration(retryCnt)) // Exponential backoff
			continue
		}
		for i, result := range results {
			testId := fmt.Sprintf("%d", i+1)
			switch result.Status {
			case "OK":
				logrus.Infof("Evaluation successful. Output:\n %s", result.Output)
				db.DB.UpdateTestStatus(s, testId , models.OK)
			case "memorylimiterror":
				logrus.Errorf("Memory limit exceeded. Output:\n %s", result.Output)
				db.DB.UpdateTestStatus(s, testId , models.MemoryLimitError)
			default:
				logrus.Warnf("Unknown status: %s. Output:\n %s", result.Status, result.Output)
				// Handle unknown status
			}
		}
		db.DB.UpdateSubmissionStatus(s, models.Evaluated)
		break
	}
}
