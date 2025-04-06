package main

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"oj/judge/api"
	"oj/judge/internal/evaluator"
)


// TODO the "goforces" is the judge in the serve mode
//      and this package is the judge in the code-runner mode
func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{})

	logger.Debug("This is a debug message")

	port := ":8080"

	router := api.SetupRoutes()
	// Setup DB
	api.Eval = evaluator.NewMockEvaluator()

	// Start the HTTP server.
	logrus.Infof("Starting API server on port %s...\n", port)
	logrus.Fatal(http.ListenAndServe(port, router))

}
