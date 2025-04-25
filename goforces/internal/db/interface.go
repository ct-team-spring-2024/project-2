package db

import (
	"oj/goforces/internal/models"
)

var DB Database

type Database interface {
	GetUserSubmission(userID int) []models.Submission
	AddSubmission(s models.Submission) error
	GetUserByID(userID int) (*models.User, error)
	CreateUser(user models.User) error
	GetProblemByID(problemID int) (*models.Problem, error)
	CreateProblem(problem models.Problem) error
	UpdateProblemStatus(problemID int, status models.ProblemStatus) error
	GetProblems() ([]models.Problem, error)
	UpdateProblem(problemId int, newProblem models.Problem) error
	GetUsers() []models.User
	UpdateUsers(userId int, newUser models.User) error
}
