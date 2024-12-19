package ssh

import (
	"fmt"
	"github.com/ksctl/ksctl/pkg/consts"
	"sync"
)

type ExecutionPipeline interface {
	NextScript() *Script
	String() string
	Append(Script)
	IsCompleted() bool
}

type Script struct {
	Name           string
	ShellScript    string
	CanRetry       bool
	MaxRetries     uint8
	ScriptExecutor consts.KsctlSupportedScriptRunners
}

type Scripts struct {
	data    []Script
	currIdx int
	mu      *sync.Mutex
}

func NewExecutionPipeline() *Scripts {
	return &Scripts{
		mu:      &sync.Mutex{},
		currIdx: -1,
	}
}

func (s *Scripts) TestAllScripts() []Script {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.data
}

func (s *Scripts) TestLen() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.currIdx
}

func (s *Scripts) NextScript() *Script {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.data) == s.currIdx {
		return nil
	}

	data := s.data[s.currIdx]
	s.currIdx++
	return &data
}

func (s *Scripts) IsCompleted() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.data == nil {
		return false
	}
	if len(s.data) == s.currIdx {
		return true
	} else {
		return false
	}
}

func (s *Scripts) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return fmt.Sprintf("%#v", s.data)
}

func (s *Scripts) Append(script Script) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.data == nil {
		s.currIdx = 0
	}
	shebang := ""
	switch script.ScriptExecutor {
	case consts.LinuxBash:
		shebang = "#!" + string(consts.LinuxBash)
	case consts.LinuxSh:
		shebang = "#!" + string(consts.LinuxSh)
	default:
		shebang = "#!" + string(consts.LinuxBash)
	}
	script.ShellScript = shebang + "\n" + script.ShellScript
	s.data = append(s.data, script)
}
