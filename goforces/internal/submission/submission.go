package submission

import (
	"oj/goforces/internal"
	"oj/goforces/internal/db"
)

func GetUserSubmission(db db.Database, user internal.User) []internal.Submission {
	return db.GetUserSubmission(user)
}
