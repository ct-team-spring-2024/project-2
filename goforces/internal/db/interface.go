package db

import (
	"oj/goforces/internal"
)

type Database interface {
	GetUserSubmission(internal.User) []internal.Submission
}
