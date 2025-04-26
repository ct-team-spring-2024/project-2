package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"oj/goforces/internal/models"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

var conn *pgx.Conn

type postgresDB struct {
	conn *pgx.Conn
	mu   sync.Mutex
}

func ConnectToDB() *postgresDB {
	connStr := "postgres://postgres:example@localhost:5432/postgres"

	// Connect to the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		logrus.Errorf("Unable to connect to database: %v\n", err)
	}
	//defer conn.Close(ctx)
	db := postgresDB{
		conn: conn,
	}
	//-----------------For testing
	// db.AddSubmission(models.Submission{
	//	UserId:    1,
	//	ProblemId: 2,
	//	Code:      "jfksdfjskdljf",
	//	Status:    "ssss",
	// })
	// newProblem := models.Problem{
	//	OwnerId:     1, // example owner
	//	Title:       "Example Problem",
	//	Statement:   "Write a function to print 'Hello World'",
	//	TimeLimit:   3,   // in seconds
	//	MemoryLimit: 256, // in MB
	//	Input:       "No input required",
	//	Output:      "Hello World",
	//	Status:      "Draft",
	//	Feedback:    "Needs more testing",
	//	PublishDate: time.Now(), // Optional, could be `nil` if not yet published
	// }
	// db.CreateProblem(newProblem)
	// logrus.Info(db.GetProblemByID(1))
	// db.UpdateProblemStatus(1, "Published")
	// logrus.Info(db.GetProblemByID(1))
	// logrus.Info(db.GetUserSubmission(1))
	// user, err := db.GetUserByID(1)
	// logrus.Info(user)

	logrus.Info("Connected to Postgre")

	// Sample query: Get current time
	//	var now time.Time
	// err = conn.QueryRow(ctx, "SELECT NOW()").Scan(&now)
	// if err != nil {
	//	logrus.Error("Query failed: %v\n", err)
	// }

	// logrus.Info("Current time from DB: %v\n", now)
	return &db
}
func (db *postgresDB) GetUserSubmission(userID int) []models.Submission {
	db.mu.Lock()
	defer db.mu.Unlock()

	query := `
	SELECT id, user_id, problem_id, code, tests_status, submission_status
	FROM submissions
	WHERE user_id = $1
	`

	rows, err := db.conn.Query(context.Background(), query, userID)
	if err != nil {
		logrus.WithError(err).Error("failed to execute query")
		return nil
	}
	defer rows.Close()

	var submissions []models.Submission
	for rows.Next() {
		var (
			s              models.Submission
			testsStatus    []byte
			submissionStat string
		)

		err := rows.Scan(&s.ID, &s.UserId, &s.ProblemId, &s.Code, &testsStatus, &submissionStat)
		if err != nil {
			logrus.WithError(err).Error("failed to scan row")
			continue
		}
		if err := json.Unmarshal(testsStatus, &s.TestsStatus); err != nil {
			logrus.WithError(err).Error("failed to deserialize tests_status")
			continue
		}

		s.SubmissionStatus = models.SubmissionStatus(submissionStat)
		submissions = append(submissions, s)
	}

	if err := rows.Err(); err != nil {
		logrus.WithError(err).Error("error occurred during row iteration")
	}

	return submissions
}

func (db *postgresDB) GetSubmission(submissionID int) models.Submission {
	db.mu.Lock()
	defer db.mu.Unlock()

	query := `
	SELECT id, user_id, problem_id, code, tests_status, submission_status
	FROM submissions
	WHERE id = $1
	`

	var (
		submission     models.Submission
		testsStatus    []byte
		submissionStat string
	)

	err := db.conn.QueryRow(context.Background(), query, submissionID).Scan(
		&submission.ID,
		&submission.UserId,
		&submission.ProblemId,
		&submission.Code,
		&testsStatus,
		&submissionStat,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			logrus.Errorf("%v", fmt.Errorf("submission with ID %d not found", submissionID))
			return models.Submission{}
		}
		logrus.Errorf("%v", fmt.Errorf("failed to query submission: %w", err))
		return models.Submission{}
	}

	if err := json.Unmarshal(testsStatus, &submission.TestsStatus); err != nil {
		logrus.Errorf("%v", fmt.Errorf("failed to deserialize tests_status: %w", err))
		return models.Submission{}
	}

	submission.SubmissionStatus = models.SubmissionStatus(submissionStat)

	// Return the submission and no error
	return submission
}

func (db *postgresDB) AddSubmission(s models.Submission) (int, error) {
	testsStatusJSON, err := json.Marshal(s.TestsStatus)
	if err != nil {
		return 0, fmt.Errorf("failed to serialize tests status: %w", err)
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	query := `
	INSERT INTO submissions (user_id, problem_id, code, tests_status, submission_status)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id
	`

	var id int
	err = db.conn.QueryRow(context.Background(), query, s.UserId, s.ProblemId, s.Code, testsStatusJSON, s.SubmissionStatus).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert submission: %w", err)
	}

	return id, nil
}

func (db *postgresDB) UpdateSubmissionStatus(submissionID int, status models.SubmissionStatus) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	query := `
	UPDATE submissions
	SET submission_status = $1
	WHERE id = $2
	`

	_, err := db.conn.Exec(context.Background(), query, string(status), submissionID)
	if err != nil {
		logrus.WithError(err).Error("failed to update submission status")
		return fmt.Errorf("failed to update submission status: %w", err)
	}

	return nil
}

func (db *postgresDB) UpdateTestStatus(s models.Submission, testId string, testStatus models.TestStatus) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	query := `
	SELECT tests_status
	FROM submissions
	WHERE id = $1
	`

	var testsStatusJSON []byte
	err := db.conn.QueryRow(context.Background(), query, s.ID).Scan(&testsStatusJSON)
	if err != nil {
		return fmt.Errorf("failed to fetch tests_status for submission ID %d: %w", s.ID, err)
	}

	var testsStatus map[string]models.TestStatus
	if err := json.Unmarshal(testsStatusJSON, &testsStatus); err != nil {
		return fmt.Errorf("failed to deserialize tests_status for submission ID %d: %w", s.ID, err)
	}

	testsStatus[testId] = testStatus

	updatedTestsStatusJSON, err := json.Marshal(testsStatus)
	if err != nil {
		return fmt.Errorf("failed to serialize updated tests_status for submission ID %d: %w", s.ID, err)
	}

	updateQuery := `
	UPDATE submissions
	SET tests_status = $1
	WHERE id = $2
	`

	_, err = db.conn.Exec(context.Background(), updateQuery, updatedTestsStatusJSON, s.ID)
	if err != nil {
		return fmt.Errorf("failed to update tests_status for submission ID %d: %w", s.ID, err)
	}

	return nil
}

func (db *postgresDB) GetUserByID(userID int) (*models.User, error) {
	// Define the SQL query to fetch user by ID
	db.mu.Lock()
	defer db.mu.Unlock()
	query := `
		SELECT user_id, username, email, password, role
		FROM users
		WHERE user_id = $1
	`

	// Prepare a variable to hold the result
	var user models.User

	// Execute the query and scan the result into the user struct
	err := db.conn.QueryRow(context.Background(), query, userID).Scan(&user.UserId, &user.Username, &user.Email, &user.Password, &user.Role)
	if err != nil {
		// Check if the error is due to no rows being found
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	// Return the user and no error
	return &user, nil
}
func (db *postgresDB) CreateUser(user models.User) (int ,error) {
	query := `
	INSERT INTO users (username, email, password, role)
	VALUES ($1, $2, $3, $4)
	RETURNING user_id
	`

	var id int
	err := db.conn.QueryRow(context.Background(), query, user.Username, user.Email, user.Password, user.Role).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return id, nil
}
func (db *postgresDB) GetProblemByID(problemID int) (*models.Problem, error) {
	// Define the SQL query to fetch the problem by its problem_id
	db.mu.Lock()
	defer db.mu.Unlock()
	query := `
	SELECT problem_id, owner_id, title, statement, time_limit, memory_limit,
	       inputs, outputs, status, feedback, publish_date
	FROM problems
	WHERE problem_id = $1
    `

	var (
		problem     models.Problem
		inputsJSON  []byte
		outputsJSON []byte
	)

	err := db.conn.QueryRow(context.Background(), query, problemID).Scan(
		&problem.ProblemId,
		&problem.OwnerId,
		&problem.Title,
		&problem.Statement,
		&problem.TimeLimit,
		&problem.MemoryLimit,
		&inputsJSON,
		&outputsJSON,
		&problem.Status,
		&problem.Feedback,
		&problem.PublishDate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("problem with ID %d not found", problemID)
		}
		return nil, fmt.Errorf("failed to query problem: %w", err)
	}

	if err := json.Unmarshal(inputsJSON, &problem.Inputs); err != nil {
		return nil, fmt.Errorf("failed to deserialize inputs: %w", err)
	}

	if err := json.Unmarshal(outputsJSON, &problem.Outputs); err != nil {
		return nil, fmt.Errorf("failed to deserialize outputs: %w", err)
	}

	return &problem, nil
}

func (db *postgresDB) CreateProblem(problem models.Problem) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	inputsJSON, err := json.Marshal(problem.Inputs)
	if err != nil {
		logrus.WithError(err).Error("failed to serialize inputs")
		return fmt.Errorf("failed to serialize inputs: %w", err)
	}
	outputsJSON, err := json.Marshal(problem.Outputs)
	if err != nil {
		logrus.WithError(err).Error("failed to serialize outputs")
		return fmt.Errorf("failed to serialize outputs: %w", err)
	}

	query := `
	INSERT INTO problems (owner_id, title, statement, time_limit, memory_limit,
			      inputs, outputs, status, feedback, publish_date)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = db.conn.Exec(context.Background(), query, problem.OwnerId, problem.Title, problem.Statement, problem.TimeLimit,
		problem.MemoryLimit, inputsJSON, outputsJSON, problem.Status, problem.Feedback, problem.PublishDate)

	if err != nil {
		logrus.WithError(err).Error("failed to insert problem")
		return fmt.Errorf("failed to insert problem: %w", err)
	}

	return nil
}
func (db *postgresDB) UpdateProblemStatus(problemID int, status models.ProblemStatus) error {
	// Define the SQL query to update the problem's status
	db.mu.Lock()
	defer db.mu.Unlock()
	query := `
		UPDATE problems
		SET status = $1
		WHERE problem_id = $2
	`

	// Execute the update query
	_, err := db.conn.Exec(context.Background(), query, status, problemID)
	if err != nil {
		return err
	}

	return nil
}
func (db *postgresDB) UpdateProblem(problemId int, newProblem models.Problem) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	inputsJSON, err := json.Marshal(newProblem.Inputs)
	if err != nil {
		logrus.WithError(err).Error("failed to serialize inputs")
		return fmt.Errorf("failed to serialize inputs: %w", err)
	}
	outputsJSON, err := json.Marshal(newProblem.Outputs)
	if err != nil {
		logrus.WithError(err).Error("failed to serialize outputs")
		return fmt.Errorf("failed to serialize outputs: %w", err)
	}

	query := `
	UPDATE problems
	SET owner_id = $1, title = $2, statement = $3, time_limit = $4, memory_limit = $5,
	    inputs = $6, outputs = $7, status = $8, feedback = $9, publish_date = $10
	WHERE problem_id = $11
	`

	_, err = db.conn.Exec(context.Background(), query, newProblem.OwnerId, newProblem.Title, newProblem.Statement,
		newProblem.TimeLimit, newProblem.MemoryLimit, inputsJSON, outputsJSON,
		newProblem.Status, newProblem.Feedback, newProblem.PublishDate, problemId)
	if err != nil {
		logrus.WithError(err).Error("failed to update problem")
		return fmt.Errorf("failed to update problem: %w", err)
	}

	return nil
}

func (db *postgresDB) GetProblems() ([]models.Problem, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Define the SQL query to get all problems
	query := `
	SELECT problem_id, owner_id, title, statement, time_limit, memory_limit,
	       inputs, outputs, status, feedback, publish_date
	FROM problems
	`
	rows, err := db.conn.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var problems []models.Problem
	for rows.Next() {
		var (
			problem     models.Problem
			inputsJSON  []byte
			outputsJSON []byte
		)

		err := rows.Scan(&problem.ProblemId, &problem.OwnerId, &problem.Title, &problem.Statement,
			&problem.TimeLimit, &problem.MemoryLimit, &inputsJSON, &outputsJSON,
			&problem.Status, &problem.Feedback, &problem.PublishDate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if err := json.Unmarshal(inputsJSON, &problem.Inputs); err != nil {
			return nil, fmt.Errorf("failed to deserialize inputs: %w", err)
		}
		if err := json.Unmarshal(outputsJSON, &problem.Outputs); err != nil {
			return nil, fmt.Errorf("failed to deserialize outputs: %w", err)
		}
		problems = append(problems, problem)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}

	return problems, nil
}
func (db *postgresDB) GetUsers() []models.User {
	// Define the SQL query to retrieve all users
	db.mu.Lock()
	defer db.mu.Unlock()
	query := `
		SELECT user_id, username, email, password, role
		FROM users
	`

	// Execute the query and get the rows
	rows, err := db.conn.Query(context.Background(), query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	// Create a slice to hold the users
	var users []models.User

	// Iterate over the rows and scan each result into the users slice
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.UserId, &user.Username, &user.Email, &user.Password, &user.Role)
		if err != nil {
			return nil
		}
		users = append(users, user)
	}

	// Check for any error during iteration
	if err := rows.Err(); err != nil {
		return nil
	}

	// Return the slice of users
	return users
}
func (db *postgresDB) UpdateUsers(userId int, newUser models.User) error {
	// Define the SQL query to update the user details
	query := `
		UPDATE users
		SET username = $1, email = $2, password = $3, role = $4
		WHERE user_id = $5
	`

	// Execute the update query
	_, err := db.conn.Exec(context.Background(), query, newUser.Username, newUser.Email, newUser.Password, newUser.Role, userId)
	if err != nil {
		return err
	}

	return nil
}
