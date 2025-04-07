package evaluator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	// "github.com/sirupsen/logrus"
)

type DockerEvaluator struct {
}

func NewDockerEvaluator() *DockerEvaluator {
	return &DockerEvaluator{}
}

func (e *DockerEvaluator) EvalCode(code string, inputs []string, timeout time.Duration) (Result, []string) {
	// start the container with the given code
	// for each of the input:
	// wait for the result. we should get the result as soon as it is ready.
	// if the result wasn't ready, the string will be "timeout"
	// return Result{}, inputs
	// Step 1: Create a temporary directory for inputs
	inputTempDir, err := os.MkdirTemp("", "inputs")
	outputTempDir, err := os.MkdirTemp("", "outputs")
	if err != nil {
		return Result{Error: fmt.Errorf("failed to create temp dir: %v", err)}, nil
	}
	defer os.RemoveAll(inputTempDir)
	// defer os.RemoveAll(outputTempDir)

	// Step 2: Create input files in the temporary directory
	for i, input := range inputs {
		filePath := filepath.Join(inputTempDir, fmt.Sprintf("%d.txt", i+1))
		err := os.WriteFile(filePath, []byte(input), 0644)
		if err != nil {
			return Result{Error: fmt.Errorf("failed to write input file: %v", err)}, nil
		}
	}

	// Step 3: Start the Docker container
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return Result{Error: fmt.Errorf("failed to create Docker client: %v", err)}, nil
	}

	imageName := "dockerevaluator"

	// Define the container configuration
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   []string{"./main"},
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/root/inputs", inputTempDir),
			fmt.Sprintf("%s:/root/outputs", outputTempDir),
		}, // Mount the inputs folder
	}, nil, nil, "")
	if err != nil {
		return Result{Error: fmt.Errorf("failed to create container: %v", err)}, nil
	}

	// Start the container
	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return Result{Error: fmt.Errorf("failed to start container: %v", err)}, nil
	}

	// Wait for the container to finish or timeout
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	// TODO Maybe we need a timeout for here too (in case of unexpected things)
	select {
	case err := <-errCh:
		logrus.Info("#1")
		if err != nil {
			return Result{Error: fmt.Errorf("container failed: %v", err)}, nil
		}
	case <-statusCh:
		// Container finished successfully
		logrus.Info("#2")
	}
	PrintContentsOfTxtFiles(outputTempDir)

	// Remove the container
	// err = cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{})
	// if err != nil {
	//	return Result{Error: fmt.Errorf("failed to remove container: %v", err)}, nil
	// }

	// Return the result
	return Result{Output: "gooz"}, inputs
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
