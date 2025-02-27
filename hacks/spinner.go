// Copyright 2025 Ksctl Authors
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

package hacks

import (
	"fmt"
	"time"
)

const (
	RESET = "\033[0m"
	BLUE  = "\033[94m"
	GREEN = "\033[92m"
	RED   = "\033[91m"
	CYAN  = "\033[96m"
	BOLD  = "\033[1m"
)

// Spinner represents a spinning animation
type Spinner struct {
	chars     []string
	done      chan bool
	active    bool
	startTime time.Time
}

// NewSpinner creates a new spinner instance
func NewSpinner() *Spinner {
	return &Spinner{
		chars:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		done:   make(chan bool),
		active: false,
	}
}

// Start begins the spinner animation in a goroutine
func (s *Spinner) Start() {
	if s.active {
		return
	}
	s.active = true
	s.startTime = time.Now()

	go func() {
		for i := 0; ; i = (i + 1) % len(s.chars) {
			select {
			case <-s.done:
				fmt.Print("\r") // Clear the spinner
				return
			default:
				elapsed := time.Since(s.startTime).Round(time.Second)
				fmt.Printf("\r%s%s%s %s", GREEN, s.chars[i], RESET, elapsed)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

// Stop halts the spinner animation
func (s *Spinner) Stop() {
	if !s.active {
		return
	}
	s.done <- true
	s.active = false
}
