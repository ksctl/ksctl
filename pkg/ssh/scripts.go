// Copyright 2024 Ksctl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ssh

import (
	"fmt"
	"github.com/ksctl/ksctl/v2/pkg/consts"
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
