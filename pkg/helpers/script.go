package helpers

import (
	"fmt"
	"sync"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

// Scripts is goroutine safe
type Scripts struct {
	data    []resources.Script
	currIdx int
	mu      *sync.Mutex
}

func NewScriptCollection() resources.ScriptCollection {
	return &Scripts{
		mu:      &sync.Mutex{},
		currIdx: -1,
	}
}

func (s *Scripts) TestAllScripts() []resources.Script {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.data
}

func (s *Scripts) TestLen() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.currIdx
}

func (s *Scripts) NextScript() *resources.Script {
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

func (s *Scripts) Append(script resources.Script) {
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
