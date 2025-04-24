package evaluator

import "time"

type Result struct {
	Output string
	Error  error
}

type Evaluator interface {
	EvalCode(code string, inputs []string, timelimit time.Duration, memorylimit int) (Result, []string)
}
