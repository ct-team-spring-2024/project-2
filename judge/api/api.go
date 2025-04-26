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

	overallresult, results := Eval.EvalCode(evalRequest.Code, evalRequest.Inputs,
		time.Duration(evalRequest.Timelimit)*time.Millisecond,
		evalRequest.Memorylimit)
	logrus.Infof("overallresult => %+v", overallresult)
	logrus.Infof("results => %+v", results)

	// Respond with the result
	response := map[string]interface{}{
		"overallresult": overallresult,
		"results":       results,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", helloWorld)
	mux.HandleFunc("/eval-code", EvalCode)
	return mux
}
