package evaluator

import "time"

type Result struct {
	Output string
	Error  error
}

type Evaluator interface {
	EvalCode(code string, inputs []string, timeout time.Duration) (Result, []string)
}
