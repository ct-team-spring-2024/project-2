package services

import (
	"oj/goforces/internal/db"
	"oj/goforces/internal/models"
)

func CreateSubmission(userId int, problemId int, code, language string) (int, error) {
	newSub := models.Submission{
		UserId:    userId,
		ProblemId: problemId,
		Code:      code,
		TestsStatus:      make(map[string]models.TestStatus),
		SubmissionStatus: models.Submitted,
	}

	subId, err := db.DB.AddSubmission(newSub)
	if err != nil {
		return -1, err
	}
	return subId, nil
}

func GetSubmissionsByUser(userId int) ([]models.Submission, error) {
	user := models.User{UserId: userId}
	submissions := db.DB.GetUserSubmission(user.UserId)
	return submissions, nil
}
