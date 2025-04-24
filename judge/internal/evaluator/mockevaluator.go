package evaluator
import (
	"time"
)

type MockEvaluator struct {

}

func NewMockEvaluator() *MockEvaluator {
	return &MockEvaluator{}
}

func (e *MockEvaluator) EvalCode(code string, inputs []string, timelimit time.Duration, memorylimit int) (Result, []string) {
	return Result{}, inputs
}
