package services

import (
	"errors"
	"oj/goforces/internal/db"
	"oj/goforces/internal/models"
)

var submissionIdCounter = 1

func CreateSubmission(userId, problemId int, code, language string) (models.Submission, error) {
	if code == "" || language == "" {
		return models.Submission{}, errors.New("code and language cannot be empty")
	}

	newSub := models.Submission{
		ID:        submissionIdCounter,
		UserId:    userId,
		ProblemId: problemId,
		Code:      code,
		Status:    "Not Examined",
	}
	submissionIdCounter++

	err := db.DB.AddSubmission(newSub)
	if err != nil {
		return models.Submission{}, err
	}
	return newSub, nil
}

func GetSubmissionsByUser(userId int) ([]models.Submission, error) {
	user := models.User{UserId: userId}
	submissions := db.DB.GetUserSubmission(user.UserId)
	return submissions, nil
}
