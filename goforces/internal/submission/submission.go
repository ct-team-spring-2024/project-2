package submission

import (
	"oj/goforces/internal"
	"oj/goforces/internal/db"
	"oj/goforces/internal/models"
)

func GetUserSubmission(db db.Database, user internal.User) []models.Submission {
	return db.GetUserSubmission(user.UserId)
}

func AddSubmission(db db.Database, submission models.Submission) error {
	return db.AddSubmission(submission)
}
