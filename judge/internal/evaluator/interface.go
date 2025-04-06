package evaluator

import "time"

type Result struct {

}

type Evaluator interface {
	EvalCode(code string, inputs []string, timeout time.Duration) (Result, []string)
}
