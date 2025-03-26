package main

import (
	"github.com/sirupsen/logrus"
)


func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{})

	logger.Debug("This is a debug message")
}
