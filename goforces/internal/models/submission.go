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
	ID        int              `json:"id"`
	UserId    int              `json:"userId"`
	ProblemId int              `json:"problemId"`
	Code      string           `json:"code"`
	Status    SubmissionStatus `json:"status"`
}

func generateUniqueId() int {
	uuidValue := uuid.New()
	return int(uuidValue.ID())
}

func NewSubmission(userId int, problemId int, code string) Submission {
	return Submission{
		ID:        generateUniqueId(),
		UserId:    userId,
		ProblemId: problemId,
		Code:      code,
		Status:    "Submitted",
	}
}
