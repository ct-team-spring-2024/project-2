package evaluator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type DockerEvaluator struct {
}

func NewDockerEvaluator() *DockerEvaluator {
	return &DockerEvaluator{}
}

func getResult(resultFilePath string) map[string]interface{} {
	resultContent, _ := os.ReadFile(resultFilePath)
	var result map[string]interface{}
	json.Unmarshal(resultContent, &result)
	return result
	// logrus.Infof("init result => %+v", result)

}

func (e *DockerEvaluator) EvalCode(code string, inputs []string, timelimit time.Duration, memorylimit int) (OverallResult, []Result) {
	results := make([]Result, 0)
	userCodeFolderPath, err := os.MkdirTemp("", "usercode")
	userCodeFilePath := fmt.Sprintf("%s/usercode.go", userCodeFolderPath)
	err = os.WriteFile(userCodeFilePath, []byte(code), 0644)
	if err != nil {
		return OverallResult{
				Description: "Init Eval Failed",
				Error:       fmt.Errorf("failed to write usercode.go: %v", err)},
			results
	}
	defer os.RemoveAll(userCodeFolderPath)

	configFolderPath, err := os.MkdirTemp("", "config")
	configFilePath := fmt.Sprintf("%s/config.txt", configFolderPath)
	err = os.WriteFile(configFilePath, []byte(fmt.Sprintf("timeout=%d", timelimit.Milliseconds())), 0644)
	if err != nil {
		return OverallResult{
				Description: "Init Eval Failed",
				Error:       fmt.Errorf("failed to write config: %v", err)},
			results
	}
	defer os.RemoveAll(configFolderPath)

	resultFolderPath, err := os.MkdirTemp("", "result")
	resultFilePath := fmt.Sprintf("%s/result.json", resultFolderPath)
	err = os.WriteFile(resultFilePath, []byte("{}"), 0644)
	if err != nil {
		return OverallResult{
			Description: "Init Eval Failed",
			Error:       fmt.Errorf("failed to create result.json: %v", err)},
			results
	}
	defer os.RemoveAll(resultFolderPath)

	inputDir, err := os.MkdirTemp("", "inputs")
	inputFilePath := filepath.Join(inputDir, "input.txt")
	err = os.WriteFile(inputFilePath, []byte(""), 0644)
	if err != nil {
		return OverallResult{
			Description: "Init Eval Failed",
			Error:       fmt.Errorf("failed to create temp dir: %v", err)},
			results
	}
	defer os.RemoveAll(inputDir)

	outputDir, err := os.MkdirTemp("", "outputs")
	outputFilePath := filepath.Join(outputDir, "output.txt")
	os.WriteFile(outputFilePath, []byte(""), 0644)
	if err != nil {
		return OverallResult{
			Description: "Init Eval Failed",
			Error:       fmt.Errorf("failed to create temp dir: %v", err)},
			results
	}
	defer os.RemoveAll(outputDir)

	// Step 2: Start the Docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return OverallResult{
			Description: "Init Eval Failed",
			Error:       fmt.Errorf("failed to create Docker client: %v", err)},
			results
	}
	// Step 3: Create the container

	imageName := "dockerevaluator"
	logrus.Infof("ii %s \n %s \n %s \n %s", userCodeFilePath, resultFilePath, inputFilePath, outputFilePath)
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   []string{"tail", "-f", "/dev/null"},
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/app/usercode.go", userCodeFilePath),
			fmt.Sprintf("%s:/app/config.txt", configFilePath),
			fmt.Sprintf("%s:/app/result.json", resultFilePath),
			fmt.Sprintf("%s:/app/input.txt", inputFilePath),
			fmt.Sprintf("%s:/app/output.txt", outputFilePath),
		},
		Resources: container.Resources{
			Memory:     int64(memorylimit * 1024 * 1024),
			MemorySwap: int64(memorylimit * 1024 * 1024),
		},
	}, nil, nil, "")
	if err != nil {
		return OverallResult{
			Description: "Init Eval Failed",
			Error:       fmt.Errorf("failed to create container: %v", err)},
			results
	}

	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return OverallResult{
			Description: "Init Eval Failed",
			Error:       fmt.Errorf("failed to start container: %v", err)},
			results
	}

	// Run the command using docker exec
	cmd := exec.Command("docker", "exec", resp.ID, "go", "run", "/app/main.go", "--compile")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return OverallResult{
			Description: "Init Eval Failed",
			Error:       fmt.Errorf("command execution failed: %v, output: %s", err, string(output))},
			results
	}

	// Check for compile error
	result := getResult(resultFilePath)
	if result["overall_result"] == "compileerror" {
		return OverallResult{
			Description: "Compilation Failed",
			Error:       fmt.Errorf("compilation error")},
			results
	}

	// // Step 5: Run tests for each input
	for i, input := range inputs {
		testId := strconv.Itoa(i+1)
		// Write input to input.txt
		logrus.Infof("input => %s", input)
		err := os.WriteFile(inputFilePath, []byte(input), 0644)
		if err != nil {
			return OverallResult{
				Description: "Running Test Failed",
				Error:       fmt.Errorf("failed to write input file: %v", err)},
				results
		}

		cmd := exec.Command("docker", "exec", resp.ID, "go", "run", "/app/main.go", "--test-id", testId)
		cmd.CombinedOutput()

		inspectResp, inspectErr := cli.ContainerInspect(ctx, resp.ID)
		if inspectErr != nil {
			return OverallResult{
				Description: "Running Test Failed",
				Error:       fmt.Errorf("failed to inspect container: %v", inspectErr)},
				results
		}
		if inspectResp.State.OOMKilled {
			// Restart the container
			logrus.Infof("Restarting container due to OOMKilled...")
			err = cli.ContainerRestart(ctx, resp.ID, container.StopOptions{})
			if err != nil {
				return OverallResult{
					Description: "Running Test Failed",
					Error:       fmt.Errorf("failed to restart container: %v", err)},
					results
			}
		}

		result := getResult(resultFilePath)
		output := result["test_n"].(map[string]interface{})
		logrus.Infof("result => %v", result)
		logrus.Infof("output => %+v", output)
		if output[testId] == "timelimiterror" {
			results = append(results, Result{
				Status: StatusTimeLimitError,
				Output: "",
			})
		} else
		if output[testId] == "memorylimiterror" {
			results = append(results, Result{
				Status: StatusMemoryLimitError,
				Output: "",
			})
		} else
		if output[testId] == "runtimeerror" {
			results = append(results, Result{
				Status: StatusRuntimeError,
				Output: "",
			})
		} else
		if output[testId] == "ok" {
			outputContent, err := os.ReadFile(outputFilePath)
			if err != nil {
				return OverallResult{
					Description: "Running Test Failed",
					Error:       fmt.Errorf("failed to read output file for test %d: %v", i+1, err)},
					results
			}
			results = append(results, Result{
				Status: StatusOK,
				Output: string(outputContent),
			})
		}
	}

	err = cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{})
	if err != nil {
		logrus.Errorf("failed to remove container: %v", err)
	}

	// Return the final result
	return OverallResult{
		Description: "OK",
		Error:       nil,},
		results
}

func PrintContentsOfTxtFiles(folderPath string) {
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		logrus.Infof("Folder does not exist: %s", folderPath)
		return
	}

	files, err := os.ReadDir(folderPath)
	if err != nil {
		logrus.Fatalf("Failed to read directory: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".txt" {
			filePath := filepath.Join(folderPath, file.Name())

			content, err := os.ReadFile(filePath)
			if err != nil {
				logrus.Infof("Failed to read file %s: %v", filePath, err)
				continue
			}

			// Print the file contents
			fmt.Printf("Contents of file %s:\n%s\n", file.Name(), string(content))
		}
	}
}
