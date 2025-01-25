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

package test

import (
	"os"
	"testing"

	"github.com/ksctl/ksctl/hacks"
)

const (
	RED   = hacks.RED
	GREEN = hacks.GREEN
	CYAN  = hacks.CYAN
	BOLD  = hacks.BOLD
	RESET = hacks.RESET
	BLUE  = hacks.BLUE
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestUnit(t *testing.T) {
	UnitTest(t)
}

func TestIntegration(t *testing.T) {
	IntegrationTest(t)
}

func TestAll(t *testing.T) {
	UnitTest(t)
	IntegrationTest(t)
}
