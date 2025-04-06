package evaluator

import (
	"time"
	"fmt"
	"os"
	"path/filepath"
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
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
	tempDir, err := os.MkdirTemp("", "inputs")
	if err != nil {
		return Result{Error: fmt.Errorf("failed to create temp dir: %v", err)}, nil
	}
	defer os.RemoveAll(tempDir) // Clean up the temporary directory after execution

	// Step 2: Create input files in the temporary directory
	for i, input := range inputs {
		filePath := filepath.Join(tempDir, fmt.Sprintf("%d.txt", i+1))
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

	imageName := "my-go-app"

	// Define the container configuration
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   []string{"./main"},
	}, &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/root/inputs", tempDir)}, // Mount the inputs folder
	}, nil, nil, "")
	if err != nil {
		return Result{Error: fmt.Errorf("failed to create container: %v", err)}, nil
	}

	// Start the container
	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return Result{Error: fmt.Errorf("failed to start container: %v", err)}, nil
	}

	// Wait for the container to finish or timeout
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return Result{Error: fmt.Errorf("container failed: %v", err)}, nil
		}
	case <-statusCh:
		// Container finished successfully
	case <-time.After(timeout):
		// Timeout occurred
		cli.ContainerStop(ctx, resp.ID, container.StopOptions{})
		return Result{Output: "timeout"}, inputs
	}

	// Retrieve the logs from the container
	out, err = cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return Result{Error: fmt.Errorf("failed to get container logs: %v", err)}, nil
	}
	defer out.Close()

	// Read the logs
	var output strings.Builder
	_, err = io.Copy(&output, out)
	if err != nil {
		return Result{Error: fmt.Errorf("failed to read container logs: %v", err)}, nil
	}

	// Remove the container
	err = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
	if err != nil {
		return Result{Error: fmt.Errorf("failed to remove container: %v", err)}, nil
	}

	// Return the result
	return Result{Output: output.String()}, inputs
}
