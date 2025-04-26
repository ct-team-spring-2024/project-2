package db

import (
	"oj/goforces/internal/models"
)

var DB Database

type Database interface {
	GetUserSubmission(userID int) []models.Submission
	AddSubmission(s models.Submission) error
	UpdateSubmissionStatus(s models.Submission, status models.SubmissionStatus)
	UpdateTestStatus(s models.Submission, testId string, testStatus models.TestStatus)
	GetUserByID(userID int) (*models.User, error)
	CreateUser(user models.User) error
	GetProblemByID(problemID int) (*models.Problem, error)
	CreateProblem(problem models.Problem) error
	UpdateProblemStatus(problemID int, status models.ProblemStatus) error
}
