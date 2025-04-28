package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type TestResult struct {
	OverallResult string            `json:"overall_result"`
	Tests         map[string]string `json:"test_n"`
}

func readTimeout(configFile string) int {
	file, err := os.Open(configFile)
	if err != nil {
		return -1
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "timeout=") {
			value := strings.TrimPrefix(line, "timeout=")
			result, _ := strconv.Atoi(value)
			return result
		}
	}
	return -1
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

func readResultFromJSON() TestResult {
	data, err := os.ReadFile("result.json")
	if err != nil {
		logrus.Error("failed to read file: %v", err)
	}

	var result TestResult
	err = json.Unmarshal(data, &result)
	if err != nil {
		logrus.Error("failed to parse JSON: %v", err)
	}

	return result
}

func runTest(inputFile string, outputFile string, timeoutSeconds int) string {
	logrus.Infof("IN OU => %s %s", inputFile, outputFile)
	inFile, err := os.Open(inputFile)
	outFile, err := os.Create(outputFile)
	if err != nil {
		logrus.Errorf("Error creating output file %s: %v", outputFile, err)
		return "runtimeerror"
	}
	defer outFile.Close()

	// Run usercode binary with input file as stdin and timeout
	err = os.Chmod("./usercode", 0755) // Set executable permission
	if err != nil {
		logrus.Errorf("Error setting execute permission for usercode: %v", err)
		return "runtimeerror"
	}
	cmd := exec.Command("./usercode")
	cmd.Stdin = inFile
	cmd.Stdout = outFile

	// Start the command with timeout
	err = cmd.Start()
	if err != nil {
		logrus.Errorf("Error starting usercode for input %s: %v", inputFile, err)
		return "runtimeerror"
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
				return "memorylimiterror"
			}
			return "runtimeerror"
		}
		logrus.Infof("Successfully processed input %s to %s", inputFile, outputFile)
		return "ok"
	case <-time.After(time.Duration(timeoutSeconds) * time.Millisecond):
		cmd.Process.Kill()
		logrus.Errorf("Timeout for input %s after %d seconds", inputFile, timeoutSeconds)
		return "timelimiterror"
	}
}

var timeoutSeconds int = readTimeout("config.txt")

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	var compileMode bool
	var testID string
	flag.BoolVar(&compileMode, "compile", false, "Compile the user code")
	flag.StringVar(&testID, "test-id", "", "Run a single test with the given ID")
	flag.Parse()

	// Step 0: Compile user code
	if compileMode {
		logrus.Info("#1")
		result := TestResult{
			OverallResult: "OK",
			Tests:         make(map[string]string),
		}
		err := compileUserCode()
		logrus.Info("#2")
		if err != nil {
			logrus.Errorf("Compilation error: %v", err)
			result.OverallResult = "compileerror"
			writeResultToJSON(result)
		} else {
			logrus.Infof("Compilation successful")
			writeResultToJSON(result)
		}
		logrus.Info("#3")
		return
	}

	if testID != "" {
		inputFile := "input.txt"
		outputFile := "output.txt"
		result := readResultFromJSON()
		result.Tests[testID] = runTest(inputFile, outputFile, timeoutSeconds)
		writeResultToJSON(result)
		return
	}
}
