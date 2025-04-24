package evaluator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type DockerEvaluator struct {
}

func NewDockerEvaluator() *DockerEvaluator {
	return &DockerEvaluator{}
}

func (e *DockerEvaluator) EvalCode(code string, inputs []string, timelimit time.Duration, memorylimit int) (Result, []string) {
	userCodeFolderPath, err := os.MkdirTemp("", "usercode")
	userCodeFilePath := fmt.Sprintf("%s/usercode.go", userCodeFolderPath)
	err = os.WriteFile(userCodeFilePath, []byte(code), 0644)
	if err != nil {
		return Result{Error: fmt.Errorf("failed to write usercode.go: %v", err)}, nil
	}
	defer os.RemoveAll(userCodeFolderPath)

	resultFolderPath, err := os.MkdirTemp("", "result")
	resultFilePath := fmt.Sprintf("%s/result.json", resultFolderPath)
	err = os.WriteFile(resultFilePath, []byte("{}"), 0644)
	if err != nil {
		return Result{Error: fmt.Errorf("failed to create result.json: %v", err)}, nil
	}
	defer os.RemoveAll(resultFolderPath)

	inputDir, err := os.MkdirTemp("", "inputs")
	inputFilePath := filepath.Join(inputDir, "input.txt")
	err = os.WriteFile(inputFilePath, []byte(""), 0644)
	if err != nil {
		return Result{Error: fmt.Errorf("failed to create temp dir: %v", err)}, nil
	}
	defer os.RemoveAll(inputDir)

	outputDir, err := os.MkdirTemp("", "outputs")
	outputFilePath := filepath.Join(outputDir, "output.txt")
	os.WriteFile(outputFilePath, []byte(""), 0644)
	if err != nil {
		return Result{Error: fmt.Errorf("failed to create temp dir: %v", err)}, nil
	}
	defer os.RemoveAll(outputDir)

	// Step 2: Start the Docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return Result{Error: fmt.Errorf("failed to create Docker client: %v", err)}, nil
	}

	imageName := "dockerevaluator"

	// Step 3: Create the container
	logrus.Infof("ii %s \n %s \n %s \n %s", userCodeFilePath, resultFilePath, inputFilePath, outputFilePath)
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   []string{"tail", "-f", "/dev/null"},
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/app/usercode.go", userCodeFilePath),
			fmt.Sprintf("%s:/app/result.json", resultFilePath),
			fmt.Sprintf("%s:/app/input.txt", inputFilePath),
			fmt.Sprintf("%s:/app/output.txt", outputFilePath),
		},
		Resources: container.Resources{
			Memory: int64(memorylimit * 1024 * 1024),
		},
	}, nil, nil, "")
	if err != nil {
		return Result{Error: fmt.Errorf("failed to create container: %v", err)}, nil
	}

	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return Result{Error: fmt.Errorf("failed to start container: %v", err)}, nil
	}

	// Run the command using docker exec
	cmd := exec.Command("docker", "exec", resp.ID, "go", "run", "/app/main.go", "--compile")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Error: fmt.Errorf("command execution failed: %v, output: %s", err, string(output))}, nil
	}

	// Read compile_result.json
	resultContent, err := os.ReadFile(resultFilePath)
	if err != nil {
		return Result{Error: fmt.Errorf("failed to read compile_result.json: %v", err)}, nil
	}

	var result map[string]interface{}
	err = json.Unmarshal(resultContent, &result)
	logrus.Infof("compileResult => %+v", result)
	if err != nil {
		return Result{Error: fmt.Errorf("failed to parse compile_result.json: %v", err)}, nil
	}

	// Check overall_result
	if result["overall_result"] == "compileerror" {
		return Result{Error: fmt.Errorf("compilation failed")}, nil
	}

	// // Step 5: Run tests for each input
	var aggregatedResults []string
	for i, input := range inputs {
		// Write input to input.txt
		logrus.Infof("input => %s", input)
		err := os.WriteFile(inputFilePath, []byte(input), 0644)
		if err != nil {
			return Result{Error: fmt.Errorf("failed to write input file: %v", err)}, nil
		}


		cmd := exec.Command("docker", "exec", resp.ID, "go", "run", "/app/main.go", "--test-id", strconv.Itoa(i+1))
		output, err := cmd.CombinedOutput()

		inspectResp, inspectErr := cli.ContainerInspect(ctx, resp.ID)
		if inspectErr != nil {
			return Result{Error: fmt.Errorf("failed to inspect container: %v", inspectErr)}, nil
		}

		if inspectResp.State.OOMKilled {
			aggregatedResults = append(aggregatedResults, fmt.Sprintf("Test %d: Memory limit exceeded", i+1))
			continue
		}

		if err != nil {
			aggregatedResults = append(aggregatedResults, fmt.Sprintf("Test %d: Execution failed: %v, output: %s", i+1, err, string(output)))
			continue
		}

		outputContent, err := os.ReadFile(outputFilePath)
		if err != nil {
			return Result{Error: fmt.Errorf("failed to read output file for test %d: %v", i+1, err)}, nil
		}
		aggregatedResults = append(aggregatedResults, fmt.Sprintf("Test %d: %s", i+1, string(outputContent)))
	}

	// err = cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{})
	// if err != nil {
	//	logrus.Errorf("failed to remove container: %v", err)
	// }

	// Return the final result
	resultContent, err = os.ReadFile(resultFilePath)
	if err != nil {
		return Result{Error: fmt.Errorf("failed to read compile_result.json: %v", err)}, nil
	}
	err = json.Unmarshal(resultContent, &result)
	logrus.Infof("compileResult => %+v", result)

	return Result{Output: "Evaluation completed"}, aggregatedResults
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
