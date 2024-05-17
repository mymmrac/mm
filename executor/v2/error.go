package executor

import "fmt"

type ExprError struct {
	Message string
	Loc     Location
}

func NewExprError(text string, loc Location) *ExprError {
	return &ExprError{
		Message: text,
		Loc:     loc,
	}
}

func (e *ExprError) Error() string {
	if e.Loc.Size() == 1 {
		return fmt.Sprintf("expression at [%d]: %s", e.Loc.Start+1, e.Message)
	}
	return fmt.Sprintf("expression in rage [%d, %d]: %s", e.Loc.Start+1, e.Loc.End, e.Message)
}
