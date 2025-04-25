package db

import (
	"context"
	"database/sql"
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
		logrus.Error("Unable to connect to database: %v\n", err)
	}
	//defer conn.Close(ctx)
	db := postgresDB{
		conn: conn,
	}
	// db.AddSubmission(models.Submission{
	// 	UserId:    1,
	// 	ProblemId: 2,
	// 	Code:      "jfksdfjskdljf",
	// 	Status:    "ssss",
	// })
	// newProblem := models.Problem{
	// 	OwnerId:     1, // example owner
	// 	Title:       "Example Problem",
	// 	Statement:   "Write a function to print 'Hello World'",
	// 	TimeLimit:   3,   // in seconds
	// 	MemoryLimit: 256, // in MB
	// 	Input:       "No input required",
	// 	Output:      "Hello World",
	// 	Status:      "Draft",
	// 	Feedback:    "Needs more testing",
	// 	PublishDate: time.Now(), // Optional, could be `nil` if not yet published
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
	// 	logrus.Error("Query failed: %v\n", err)
	// }

	// logrus.Info("Current time from DB: %v\n", now)
	return &db
}
func (db *postgresDB) GetUserSubmission(userID int) []models.Submission {
	db.mu.Lock()
	defer db.mu.Unlock()
	rows, err := db.conn.Query(context.Background(), `
	SELECT id, user_id, problem_id, code, status
	FROM submissions
	WHERE user_id = $1
`, userID)
	if err != nil {
		logrus.Error("query error:", err)
		return nil
	}
	defer rows.Close()

	var submissions []models.Submission

	for rows.Next() {
		var s models.Submission
		err := rows.Scan(&s.ID, &s.UserId, &s.ProblemId, &s.Code, &s.Status)
		if err != nil {
			logrus.Error("scan error:", err)
			continue
		}
		submissions = append(submissions, s)
	}

	return submissions
}
func (db *postgresDB) AddSubmission(s models.Submission) error {
	var id int
	db.mu.Lock()
	defer db.mu.Unlock()
	query := `
		INSERT INTO submissions (user_id, problem_id, code, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := db.conn.QueryRow(context.Background(), query, s.UserId, s.ProblemId, s.Code, s.Status).Scan(&id)
	if err != nil {
		return err
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
func (db *postgresDB) CreateUser(user models.User) error {
	// Define the SQL query to insert a new user
	query := `
		INSERT INTO users (username, email, password, role)
		VALUES ($1, $2, $3, $4)
	`

	// Execute the insert query
	_, err := db.conn.Exec(context.Background(), query, user.Username, user.Email, user.Password, user.Role)
	if err != nil {
		return err
	}

	return nil
}
func (db *postgresDB) GetProblemByID(problemID int) (*models.Problem, error) {
	// Define the SQL query to fetch the problem by its problem_id
	db.mu.Lock()
	defer db.mu.Unlock()
	query := `
		SELECT problem_id, owner_id, title, statement, time_limit, memory_limit, 
		       input, output, status, feedback, publish_date
		FROM problems
		WHERE problem_id = $1
	`

	// Prepare a variable to hold the result
	var problem models.Problem

	// Execute the query and scan the result into the problem struct
	err := db.conn.QueryRow(context.Background(), query, problemID).Scan(&problem.ProblemId, &problem.OwnerId, &problem.Title,
		&problem.Statement, &problem.TimeLimit, &problem.MemoryLimit, &problem.Input,
		&problem.Output, &problem.Status, &problem.Feedback, &problem.PublishDate)

	if err != nil {
		// Check if no rows are found
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	// Return the problem and no error
	return &problem, nil
}
func (db *postgresDB) CreateProblem(problem models.Problem) error {
	// Define the SQL query to insert a new problem
	db.mu.Lock()
	defer db.mu.Unlock()
	query := `
		INSERT INTO problems (owner_id, title, statement, time_limit, memory_limit, 
		                       input, output, status, feedback, publish_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	// Execute the insert query
	_, err := db.conn.Exec(context.Background(), query, problem.OwnerId, problem.Title, problem.Statement, problem.TimeLimit,
		problem.MemoryLimit, problem.Input, problem.Output, problem.Status, problem.Feedback, problem.PublishDate)

	if err != nil {
		logrus.Error(err)
		return err
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
	// Define the SQL query to update the entire problem
	query := `
		UPDATE problems
		SET owner_id = $1, title = $2, statement = $3, time_limit = $4, memory_limit = $5,
		    input = $6, output = $7, status = $8, feedback = $9, publish_date = $10
		WHERE problem_id = $11
	`

	// Execute the update query
	_, err := db.conn.Exec(context.Background(), query, newProblem.OwnerId, newProblem.Title, newProblem.Statement,
		newProblem.TimeLimit, newProblem.MemoryLimit, newProblem.Input, newProblem.Output,
		newProblem.Status, newProblem.Feedback, newProblem.PublishDate, problemId)

	if err != nil {
		return err
	}

	return nil
}

func (db *postgresDB) GetProblems() ([]models.Problem, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	// Define the SQL query to get all problems
	query := `
		SELECT problem_id, owner_id, title, statement, time_limit, memory_limit, 
		       input, output, status, feedback, publish_date
		FROM problems
	`

	// Execute the query and get the rows
	rows, err := db.conn.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Create a slice to hold the problems
	var problems []models.Problem

	// Iterate over the rows and scan the result into the problems slice
	for rows.Next() {
		var problem models.Problem
		err := rows.Scan(&problem.ProblemId, &problem.OwnerId, &problem.Title, &problem.Statement,
			&problem.TimeLimit, &problem.MemoryLimit, &problem.Input, &problem.Output,
			&problem.Status, &problem.Feedback, &problem.PublishDate)
		if err != nil {
			return nil, err
		}
		problems = append(problems, problem)
	}

	// Check for any error during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Return the slice of problems
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
