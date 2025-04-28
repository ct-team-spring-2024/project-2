package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
	"bytes"
)

func genCode() string {
	return `package main

import (
	"bufio"
	"fmt"
	"os"

	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	line := strings.TrimSpace(scanner.Text())
	parts := strings.Split(line, " ")

	if len(parts) != 3 {
		fmt.Println("Error: Please provide exactly 3 numbers separated by spaces")
		return
	}

	num1, _ := strconv.Atoi(parts[0])
	num2, _ := strconv.Atoi(parts[1])

	sum := num1 + num2

	fmt.Printf("%d", sum)
}`
}

func calculatePercentile(durations []time.Duration, percentile float64) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	index := int(float64(len(durations)-1) * percentile / 100)
	return durations[index]
}

func writePercentilesToFile(outputName string, p50, p90, p99 time.Duration) {
	file, err := os.Create(outputName)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create output file %s: %v", outputName, err))
		return
	}
	defer file.Close()

	_, err = file.WriteString("Percentile,Execution Time\n")
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to write to output file %s: %v", outputName, err))
		return
	}

	_, err = file.WriteString(fmt.Sprintf("50th,%v\n", p50))
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to write to output file %s: %v", outputName, err))
		return
	}

	_, err = file.WriteString(fmt.Sprintf("90th,%v\n", p90))
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to write to output file %s: %v", outputName, err))
		return
	}

	_, err = file.WriteString(fmt.Sprintf("99th,%v\n", p99))
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to write to output file %s: %v", outputName, err))
		return
	}
}

func writeLatenciesToFile(latencyFileName string, durations []time.Duration) {
	file, err := os.Create(latencyFileName)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create latency file %s: %v", latencyFileName, err))
		return
	}
	defer file.Close()

	_, err = file.WriteString("Invocation Latency (ms)\n")
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to write to latency file %s: %v", latencyFileName, err))
		return
	}

	for _, duration := range durations {
		_, err = file.WriteString(fmt.Sprintf("%v\n", duration.Milliseconds()))
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to write to latency file %s: %v", latencyFileName, err))
			return
		}
	}

	slog.Info(fmt.Sprintf("Latencies written to file: %s", latencyFileName))
}

func callHttp(wg *sync.WaitGroup, durations *[]time.Duration, mu *sync.Mutex) {
	defer wg.Done()

	slog.Info("Invocation started!")
	startTime := time.Now()

	// Define the request payload
	payload := map[string]interface{}{
		"problemId": 178,
		"code":      genCode(),
		"language":  "go",
	}

	// Convert the payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Error marshalling payload", "error", err)
		return
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "http://localhost:8080/submit?sync=true", bytes.NewBuffer(jsonPayload))
	// req, err := http.NewRequest("POST", "http://localhost:8080/submit", bytes.NewBuffer(jsonPayload))
	if err != nil {
		slog.Error("Error creating HTTP request", "error", err)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDU5MjEzODYsInN1YiI6IjQwOCJ9.bGJl8Q_MsDLo-u1ZPWvaFPzwNwKxnsznlTLFrkgiVxc")

	// Perform the HTTP request
	client := &http.Client{
		Timeout: 30 * time.Second, // Increase the timeout to 30 seconds
	}
	client.Do(req)

	// Measure execution time
	executionTime := time.Since(startTime)

	// Append the execution time to the durations slice (thread-safe)
	mu.Lock()
	*durations = append(*durations, executionTime)
	mu.Unlock()

	slog.Info(fmt.Sprintf("Successfully invoked Lambda function: Execution Time: %v", executionTime))
}

func main() {
	var outputName string
	var latencyFileName string
	var rate float64
	var numInvocations int
	var logLevel string

	flag.StringVar(&outputName, "outputName", "result.txt", "Name of the output file")
	flag.StringVar(&latencyFileName, "latencyFile", "latencies.csv", "Name of the latency output file")
	flag.Float64Var(&rate, "rate", 10, "Rate of invocations per second")
	flag.IntVar(&numInvocations, "numInvocations", 100, "Number of invocations")
	flag.StringVar(&logLevel, "log", "info", "Log level")
	flag.Parse()

	conn := ConnectToDB()
	defer conn.Close(context.Background())

	InsertUsers(conn)

	CreateProblems(conn)

	// Logging
	var opts *slog.HandlerOptions
	if logLevel == "error" {
		opts = &slog.HandlerOptions{
			Level: slog.LevelWarn,
		}
	} else if logLevel == "info" {
		opts = &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	interval := time.Duration(float64(time.Second) / rate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var wg sync.WaitGroup

	var durations []time.Duration
	var mu sync.Mutex

	for i := 0; i < numInvocations; i++ {
		<-ticker.C

		wg.Add(1)
		// go invokeLambda(lambdaClient, functionName, numCalls, &wg, &durations, &mu)
		go callHttp(&wg, &durations, &mu)

	}

	wg.Wait()

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	p50 := calculatePercentile(durations, 50)
	p90 := calculatePercentile(durations, 90)
	p99 := calculatePercentile(durations, 99)

	slog.Info(fmt.Sprintf("50th Percentile Execution Time: %v", p50))
	slog.Info(fmt.Sprintf("90th Percentile Execution Time: %v", p90))
	slog.Info(fmt.Sprintf("99th Percentile Execution Time: %v", p99))

	fmt.Printf("dur %+v", durations)
	writePercentilesToFile(outputName, p50, p90, p99)
	writeLatenciesToFile(latencyFileName, durations)

	slog.Info("All Lambda invocations completed.")
}

func ConnectToDB() *pgx.Conn {
	connStr := "postgres://postgres:example@localhost:5432/postgres"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		logrus.Errorf("Unable to connect to database: %v\n", err)
	}
	logrus.Info("Connected to Postgre")
	return conn
}

func InsertUsers(conn *pgx.Conn) {
	ctx := context.Background()
	for i := 1; i <= 200; i++ {
		userID := i
		query := `INSERT INTO users (username, email, password, role) VALUES ($1, $2, $3, $4)`
		username := fmt.Sprintf("user%d", userID)
		email := fmt.Sprintf("user%d@example.com", userID)

		_, err := conn.Exec(ctx, query, username, email, "", "user")
		if err != nil {
			logrus.Errorf("Error inserting user %d: %v", userID, err)
			return
		}
	}
	logrus.Info("Inserted 200 users successfully.")
}

func CreateProblems(conn *pgx.Conn) {
	ctx := context.Background()
	problems := []struct {
		title       string
		statement   string
		timeLimit   int
		memoryLimit int
		inputs      string
		outputs     string
		status      string
	}{
		{"Problem 1", "Statement for Problem 1", 2, 256, `["input1"]`, `["output1"]`, "Published"},
		{"Problem 2", "Statement for Problem 2", 3, 512, `["input2"]`, `["output2"]`, "Published"},
		{"Problem 3", "Statement for Problem 3", 1, 128, `["input3"]`, `["output3"]`, "Published"},
		{"Problem 4", "Statement for Problem 4", 4, 1024, `["input4"]`, `["output4"]`, "Published"},
		{"Problem 5", "Statement for Problem 5", 2, 256, `["input5"]`, `["output5"]`, "Published"},
	}

	for _, problem := range problems {
		query := `
	    INSERT INTO problems (owner_id, title, statement, time_limit, memory_limit, inputs, outputs, status)
	    VALUES ($1, $2, $3, $4, $5, $6::JSONB, $7::JSONB, $8)
	`
		_, err := conn.Exec(ctx, query, 1, problem.title, problem.statement, problem.timeLimit, problem.memoryLimit, problem.inputs, problem.outputs, problem.status)
		if err != nil {
			logrus.Errorf("Error inserting problem: %v", err)
			return
		}
	}
	logrus.Info("Created 5 problems successfully.")
}

func Cleanup(conn *pgx.Conn) {
	ctx := context.Background()
	_, err := conn.Exec(ctx, "DELETE FROM users")
	if err != nil {
		logrus.Errorf("Error cleaning up users: %v", err)
	}
	_, err = conn.Exec(ctx, "DELETE FROM problems")
	if err != nil {
		logrus.Errorf("Error cleaning up problems: %v", err)
	}
	logrus.Info("Database cleanup completed.")
}
