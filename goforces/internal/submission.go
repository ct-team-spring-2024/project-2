package internal

import (
	"github.com/google/uuid"
)

type SubmissionStatus int
const (
	Submitted SubmissionStatus = iota
	OK
	WrongAnswer
	CompileError
	MemoryLimitError
	TimeLimitError
	RuntimeErrorError
)

type Submission struct {
	SubmissionId int64
	UserId       int64
	ProblemId    int64
	Status       SubmissionStatus
}

func generateUniqueId() int {
	uuidValue := uuid.New()
	return int(uuidValue.ID())
}

func NewSubmission(userId int64, problemId int64) *Submission {
	return &Submission{
		SubmissionId: int64(generateUniqueId()),
		UserId: userId,
		ProblemId: problemId,
		Status: Submitted,
	}
}
