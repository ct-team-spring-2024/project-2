package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"oj/judge/internal/evaluator"

	"github.com/sirupsen/logrus"
)

var Eval evaluator.Evaluator

func helloWorld(w http.ResponseWriter, r *http.Request) {
	response := "gooz"
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func EvalCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var evalRequest EvalCodeRequest
	if err := json.Unmarshal(body, &evalRequest); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if evalRequest.Code == "" {
		http.Error(w, "Code field is required", http.StatusBadRequest)
		return
	}
	if evalRequest.Timeout <= 0 {
		http.Error(w, "Timeout must be a positive integer", http.StatusBadRequest)
		return
	}

	// Simulate code evaluation with a timeout
	result, outputs := Eval.EvalCode(evalRequest.Code, evalRequest.Inputs, time.Duration(evalRequest.Timeout) * time.Millisecond)
	logrus.Infof("result and outputs => %v %v", result, outputs)

	// Respond with the result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(outputs)
}

func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", helloWorld)
	mux.HandleFunc("/eval-code", EvalCode)
	return mux
}
