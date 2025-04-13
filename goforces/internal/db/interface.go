package db

import (
	"oj/goforces/internal"
	"oj/goforces/internal/models"
)

type Database interface {
	GetUserSubmission(internal.User) []models.Submission
	AddSubmission(models.Submission) error
}
