package models

import (
	"github.com/google/uuid"
)

type TestStatus string
const (
	Unknown           TestStatus = "Unknown"
	OK                TestStatus = "OK"
	WrongAnswer       TestStatus = "WrongAnswer"
	CompileError      TestStatus = "CompileError"
	MemoryLimitError  TestStatus = "MemoryLimitError"
	TimeLimitError    TestStatus = "TimeLimitError"
	RuntimeErrorError TestStatus = "RuntimeErrorError"
)

type SubmissionStatus string
const (
	Submitted SubmissionStatus = "Submitted"
	Evaluated SubmissionStatus = "Evaluated"
)


type Submission struct {
	ID                int                         `json:"id"`
	UserId            int                         `json:"userId"`
	ProblemId         int                         `json:"problemId"`
	Code              string                      `json:"code"`
	TestsStatus       map[string]TestStatus       `json:"testsstatus"`
	SubmissionStatus  SubmissionStatus            `json:"submissionstatus"`
}

func generateUniqueId() int {
	uuidValue := uuid.New()
	return int(uuidValue.ID())
}

func NewSubmission(userId int, problemId int, code string) Submission {
	return Submission{
		ID:               generateUniqueId(),
		UserId:           userId,
		ProblemId:        problemId,
		Code:             code,
		SubmissionStatus: Submitted,
		TestsStatus:      make(map[string]TestStatus),
	}
}
