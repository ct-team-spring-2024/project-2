package evaluator

import "time"

type Status string

const (
	StatusOK               Status = "OK"
	StatusRuntimeError     Status = "runtimeerror"
	StatusMemoryLimitError Status = "memorylimiterror"
	StatusTimeLimitError   Status = "time-limiterror"
)

type Result struct {
	Status Status
	Output string
}

type OverallResult struct {
	Description string
	Error       error
}

type Evaluator interface {
	EvalCode(code string, inputs []string, timelimit time.Duration, memorylimit int) (OverallResult, []Result)
}
