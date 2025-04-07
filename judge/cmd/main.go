package main

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"oj/judge/api"
	"oj/judge/internal/evaluator"
)


// TODO the "goforces" is the judge in the serve mode
//      and "judge" is the judge in the code-runner mode
func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})

	logrus.Debug("This is a debug message")

	port := ":8080"

	router := api.SetupRoutes()
	// Setup DB
	// api.Eval = evaluator.NewMockEvaluator()
	api.Eval = evaluator.NewDockerEvaluator()

	// Start the HTTP server.
	logrus.Infof("Starting API server on port %s...\n", port)
	logrus.Fatal(http.ListenAndServe(port, router))

}
