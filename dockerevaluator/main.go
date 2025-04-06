package main

import (
	"bufio"
	"fmt"
	"io"
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
	// Step 1: Read the timeout value from the config file
	configFile := "config.txt"
	timeout, err := readTimeout(configFile)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return
	}

	// Step 2: Get all input files from the "inputs" directory
	inputDir := "inputs"
	inputFiles, err := filepath.Glob(filepath.Join(inputDir, "*.txt"))
	if err != nil {
		fmt.Println("Error reading input files:", err)
		return
	}

	// Step 3: Process each input file
	for _, inputFile := range inputFiles {
		// Extract the file ID (e.g., "1.txt" -> "1")
		fileID := strings.TrimSuffix(filepath.Base(inputFile), ".txt")
		logrus.Infof("starting input file %s", fileID)
		// Redirect stdin to the current input file
		stdinFile, err := os.Open(inputFile)
		if err != nil {
			logrus.Errorf("Error opening file %s: %v", inputFile, err)
			continue
		}
		defer stdinFile.Close()

		// Redirect os.Stdin to the opened file
		oldStdin := os.Stdin
		os.Stdin = stdinFile
		defer func() { os.Stdin = oldStdin }() // Restore original stdin later

		// Capture stdout for the run function
		r, w, err := os.Pipe()
		if err != nil {
			fmt.Printf("Error creating pipe for file %s: %v\n", inputFile, err)
			continue
		}

		// Save the original stdout and restore it later
		oldStdout := os.Stdout
		os.Stdout = w

		// Run the function with a timeout
		done := make(chan bool)
		go func() {
			run()
			done <- true
		}()

		logrus.Infof("starting input file %s #2", fileID)
		select {
		case <-done:
			// Function completed within the timeout
			w.Close()
			os.Stdout = oldStdout

			// Read the captured stdout
			var output strings.Builder
			io.Copy(&output, r)
			fmt.Printf("Output for file %s:\n%s\n", fileID, output.String())

		case <-time.After(time.Duration(timeout) * time.Millisecond):
			// Timeout occurred
			logrus.Infof("Timeout for file %s\n", fileID)
			w.Close()
			os.Stdout = oldStdout
		}
		logrus.Infof("starting input file %s #3", fileID)
	}
}
