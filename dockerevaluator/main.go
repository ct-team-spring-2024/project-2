package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
)

type TestResult struct {
	OverallResult string            `json:"overall_result"`
	Tests         map[string]string `json:"test_n"`
}

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

func compileUserCode() error {
	cmd := exec.Command("go", "build", "-o", "usercode", "usercode.go")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compilation failed: %s", string(output))
	}
	return nil
}

func writeResultToJSON(result TestResult) {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		logrus.Errorf("Error marshaling JSON: %v", err)
		return
	}

	err = os.WriteFile("result.json", jsonData, 0644)
	if err != nil {
		logrus.Errorf("Error writing JSON file: %v", err)
		return
	}
	logrus.Infof("Results written to result.json")
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})

	result := TestResult{
		OverallResult: "OK",
		Tests:         make(map[string]string),
	}

	// Step 0: Compile user code
	err := compileUserCode()
	if err != nil {
		logrus.Errorf("Compilation error: %v", err)
		result.OverallResult = "compileerror"
		// Write result to JSON file
		writeResultToJSON(result)
		return
	}

	// Step 1: Read the timeout value from the config file
	configFile := "config.txt"
	timeoutSeconds, err := readTimeout(configFile)
	if err != nil {
		logrus.Errorf("Error reading config file: %v", err)
		result.OverallResult = "runtimeerror"
		writeResultToJSON(result)
		return
	}

	// Step 2: Get all input files from the "inputs" directory
	inputDir := "inputs"
	outputDir := "outputs"
	inputFiles, err := filepath.Glob(filepath.Join(inputDir, "*.txt"))
	if err != nil {
		logrus.Errorf("Error reading input files: %v", err)
		result.OverallResult = "runtimeerror"
		writeResultToJSON(result)
		return
	}

	// Step 3: Process each input file
	logrus.Infof("inputFiles => %v", inputFiles)
	for _, inputFile := range inputFiles {
		// Extract the file ID (e.g., "1.txt" -> "1")
		fileID := strings.TrimSuffix(filepath.Base(inputFile), ".txt")
		outputFile := fmt.Sprintf("%s/out_%s.txt", outputDir, fileID)
		logrus.Infof("starting input file %s", fileID)

		// Open input file
		inFile, err := os.Open(inputFile)
		if err != nil {
			logrus.Errorf("Error opening input file %s: %v", inputFile, err)
			result.Tests[fileID] = "runtimeerror"
			continue
		}
		defer inFile.Close()

		// Create output file
		outFile, err := os.Create(outputFile)
		if err != nil {
			logrus.Errorf("Error creating output file %s: %v", outputFile, err)
			result.Tests[fileID] = "runtimeerror"
			continue
		}
		defer outFile.Close()

		// Run usercode binary with input file as stdin and timeout
		cmd := exec.Command("./usercode")
		cmd.Stdin = inFile
		cmd.Stdout = outFile

		// Start the command with timeout
		err = cmd.Start()
		if err != nil {
			logrus.Errorf("Error starting usercode for input %s: %v", inputFile, err)
			result.Tests[fileID] = "runtimeerror"
			continue
		}

		// Create a channel to signal timeout
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		// Wait for either completion or timeout
		select {
		case err := <-done:
			if err != nil {
				logrus.Errorf("Error running usercode for input %s: %v", inputFile, err)
				if strings.Contains(err.Error(), "signal: killed") {
					result.Tests[fileID] = "memorylimit"
				} else {
					result.Tests[fileID] = "runtimeerror"
				}
			} else {
				result.Tests[fileID] = "ok"
				logrus.Infof("Successfully processed input %s to %s", inputFile, outputFile)
			}
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			cmd.Process.Kill()
			logrus.Errorf("Timeout for input %s after %d seconds", inputFile, timeoutSeconds)
			result.Tests[fileID] = "timelimiterror"
		}
	}

	// Write result to JSON file
	writeResultToJSON(result)
}
