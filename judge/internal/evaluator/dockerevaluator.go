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
	compilerDockerId string
	userCodeFilePath string
	userCodeExeFilePath string
	resultFilePath string
}

func NewDockerEvaluator() *DockerEvaluator {
	dockerId, userCodeFilePath, userCodeExeFilePath, resultFilePath := InitCompilerContainer()
	return &DockerEvaluator{
		compilerDockerId: dockerId,
		userCodeFilePath: userCodeFilePath,
		userCodeExeFilePath: userCodeExeFilePath,
		resultFilePath: resultFilePath,
	}
}

func InitCompilerContainer() (string, string, string, string) {
	imageName := "dockerevaluator"
	userCodeFolderPath, _ := os.MkdirTemp("", "usercode")
	userCodeFilePath := fmt.Sprintf("%s/usercode.go", userCodeFolderPath)
	err := os.WriteFile(userCodeFilePath, []byte(""), 0644)
	if err != nil {
		logrus.Fatalf("usercode file cannot be created",)
	}

	userCodeExeFolderPath, err := os.MkdirTemp("", "usercodeexe")
	userCodeExeFilePath := fmt.Sprintf("%s/usercode", userCodeExeFolderPath)
	os.WriteFile(userCodeExeFilePath, []byte(""), 0644)
	if err != nil {
		logrus.Fatalf("usercodeexe file cannot be created")
	}

	resultFolderPath, err := os.MkdirTemp("", "result")
	resultFilePath := fmt.Sprintf("%s/result.json", resultFolderPath)
	err = os.WriteFile(resultFilePath, []byte("{}"), 0644)
	if err != nil {
		logrus.Fatalf("result file cannot be created")
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	ctx := context.Background()
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   []string{"tail", "-f", "/dev/null"},
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/app/usercode.go", userCodeFilePath),
			fmt.Sprintf("%s:/app/usercode", userCodeExeFilePath),
			fmt.Sprintf("%s:/app/result.json", resultFilePath),
		},
		Resources: container.Resources{
			Memory:     int64(2000 * 1024 * 1024),
			NanoCPUs:   int64(5000000000),
		},
	}, nil, nil, "")
	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		logrus.Fatalf("compiler-container creation failed")
	}
	return resp.ID, userCodeFilePath, userCodeExeFilePath, resultFilePath
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


	err := os.WriteFile(e.userCodeFilePath, []byte(code), 0644)

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
	logrus.Infof("#1")
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   []string{"tail", "-f", "/dev/null"},
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/app/usercode", e.userCodeExeFilePath),
			fmt.Sprintf("%s:/app/config.txt", configFilePath),
			fmt.Sprintf("%s:/app/result.json", e.resultFilePath),
			fmt.Sprintf("%s:/app/input.txt", inputFilePath),
			fmt.Sprintf("%s:/app/output.txt", outputFilePath),
		},
		Resources: container.Resources{
			Memory:     int64(memorylimit * 1024 * 1024),
			MemorySwap: int64(memorylimit * 1024 * 1024),
			NanoCPUs:   int64(1000000000),
		},
	}, nil, nil, "")
	logrus.Infof("#2")
	if err != nil {
		return OverallResult{
			Description: "Init Eval Failed",
			Error:       fmt.Errorf("failed to create container: %v", err)},
			results
	}

	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	logrus.Infof("#3")
	if err != nil {
		return OverallResult{
			Description: "Init Eval Failed",
			Error:       fmt.Errorf("failed to start container: %v", err)},
			results
	}

	// Run the command using docker exec
	start := time.Now()
	cmd := exec.Command("docker", "exec", e.compilerDockerId, "go", "run", "/app/main.go", "--compile")
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)
	logrus.Infof("#4: Command completed in %v", duration)
	logrus.Infof("FFF %s \n %+v", string(output), err)

	if err != nil {
		return OverallResult{
			Description: "Init Eval Failed",
			Error:       fmt.Errorf("command execution failed: %v, output: %s", err, string(output))},
			results
	}

	// Check for compile error
	result := getResult(e.resultFilePath)
	if result["overall_result"] == "compileerror" {
		return OverallResult{
			Description: "Compilation Failed",
			Error:       fmt.Errorf("compilation error")},
			results
	}

	time.Sleep(5 * time.Second)
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

		start := time.Now()
		cmd := exec.Command("docker", "exec", resp.ID, "go", "run", "/app/main.go", "--test-id", testId)
		cmd.CombinedOutput()
		duration := time.Since(start)
		logrus.Infof("#5: Command completed in %v", duration)

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

		result := getResult(e.resultFilePath)
		logrus.Infof("result => %v", result)
		output := result["test_n"].(map[string]interface{})
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

	// err = cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
	// if err != nil {
	//	logrus.Errorf("failed to remove container: %v", err)
	// }

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
