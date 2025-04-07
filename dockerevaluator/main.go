package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// TODO the main function of user should be changed to another name
//	before running the main.
func readTimeout(configFile string) (int, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "timeout=") {
			value := strings.TrimPrefix(line, "timeout=")
			return strconv.Atoi(value)
		}
	}

	return 0, fmt.Errorf("timeout not found in config file")
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})

	logrus.Debug("This is a debug message")

	// Step 1: Read the timeout value from the config file
	configFile := "config.txt"
	timeout, err := readTimeout(configFile)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return
	}

	// Step 2: Get all input files from the "inputs" directory
	inputDir := "inputs"
	outputDir := "outputs"
	inputFiles, err := filepath.Glob(filepath.Join(inputDir, "*.txt"))
	if err != nil {
		fmt.Println("Error reading input files:", err)
		return
	}

	// Step 3: Process each input file
	logrus.Infof("inputFiles => %v", inputFiles)
	for _, inputFile := range inputFiles {
		// Extract the file ID (e.g., "1.txt" -> "1")
		fileID := strings.TrimSuffix(filepath.Base(inputFile), ".txt")
		outputFile := fmt.Sprintf("%s/out_%s.txt", outputDir, fileID)
		logrus.Infof("starting input file %s", fileID)

		// reconfigure stdin and stdout
		stdinFile, err := os.Open(inputFile)
		if err != nil {
			logrus.Errorf("Error opening file %s: %v", inputFile, err)
			continue
		}
		defer stdinFile.Close()
		oldStdin := os.Stdin
		os.Stdin = stdinFile
		defer func() { os.Stdin = oldStdin }()

		stdoutFile, err := os.Create(outputFile)
		if err != nil {
			logrus.Errorf("Error opening file %s: %v", outputFile, err)
			continue
		}
		defer stdoutFile.Close()
		oldStdout := os.Stdout
		os.Stdout = stdoutFile
		defer func() { os.Stdout = oldStdout }()

		// Run the function with a timeout
		done := make(chan bool)
		go func() {
			run()
			done <- true
		}()
		select {
		case <-done:
			logrus.Infof("Successfully executed %s", fileID)
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			logrus.Infof("Timeout for file %s", fileID)
		}
		logrus.Infof("starting input file %s #3", fileID)
	}
}
