package services

import (
	"oj/goforces/internal/db"
	"oj/goforces/internal/models"
)

type SubmissionStats struct {
	Total          int     `json:"total"`
	SuccessCount   int     `json:"successCount"`
	SuccessPercent float64 `json:"successPercent"`
	FailCount      int     `json:"failCount"`
	FailPercent    float64 `json:"failPercent"`
	ErrorCount     int     `json:"errorCount"`
	ErrorPercent   float64 `json:"errorPercent"`
}

func GetSubmissionStats(user models.User) SubmissionStats {
	submissions := db.DB.GetUserSubmission(user.UserId)
	total := len(submissions)
	var success, fail, errorCount int
	for _, sub := range submissions {
		switch sub.SubmissionStatus {
		case models.Submitted:
			success++
		default:
			errorCount++
		}
	}
	var successPercent, failPercent, errorPercent float64
	if total > 0 {
		successPercent = (float64(success) / float64(total)) * 100
		failPercent = (float64(fail) / float64(total)) * 100
		errorPercent = (float64(errorCount) / float64(total)) * 100
	}
	return SubmissionStats{
		Total:          total,
		SuccessCount:   success,
		SuccessPercent: successPercent,
		FailCount:      fail,
		FailPercent:    failPercent,
		ErrorCount:     errorCount,
		ErrorPercent:   errorPercent,
	}
}
