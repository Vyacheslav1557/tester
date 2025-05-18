package tester

import (
	"fmt"
	"github.com/Vyacheslav1557/tester/internal/models"
)

type StateErr struct {
	State models.State
	Msg   string
}

func (s *StateErr) Error() string {
	return fmt.Sprintf("state: %d, msg: %s", s.State, s.Msg)
}

var (
	CompilationErr         = &StateErr{State: models.GotCE, Msg: "compilation error"}
	TimeLimitExceededErr   = &StateErr{State: models.GotTL, Msg: "time limit exceeded error"}
	MemoryLimitExceededErr = &StateErr{State: models.GotML, Msg: "memory limit exceeded error"}
	RuntimeErr             = &StateErr{State: models.GotRE, Msg: "runtime error"}
	PresentationErr        = &StateErr{State: models.GotPE, Msg: "presentation error"}
	WrongAnswerErr         = &StateErr{State: models.GotWA, Msg: "wrong answer"}
)
