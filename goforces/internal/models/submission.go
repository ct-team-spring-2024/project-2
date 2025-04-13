package models

import (
	"github.com/google/uuid"
)

type SubmissionStatus string

const (
	Submitted         SubmissionStatus = "Submitted"
	OK                SubmissionStatus = "OK"
	WrongAnswer       SubmissionStatus = "WrongAnswer"
	CompileError      SubmissionStatus = "CompileError"
	MemoryLimitError  SubmissionStatus = "MemoryLimitError"
	TimeLimitError    SubmissionStatus = "TimeLimitError"
	RuntimeErrorError SubmissionStatus = "RuntimeErrorError"
)

type Submission struct {
	SubmissionId int              `json:"submission_id"`
	UserId       int              `json:"user_id"`
	ProblemId    int              `json:"problem_id"`
	Status       SubmissionStatus `json:"status"`
}

func generateUniqueId() int {
	uuidValue := uuid.New()
	return int(uuidValue.ID())
}

func NewSubmission(userId int, problemId int) *Submission {
	return &Submission{
		SubmissionId: int(generateUniqueId()),
		UserId:       userId,
		ProblemId:    problemId,
		Status:       Submitted,
	}
}
