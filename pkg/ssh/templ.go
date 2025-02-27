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
	"testing"

	"gotest.tools/v3/assert"
)

func HelperTestTemplate(t *testing.T, testData []Script, f func() ExecutionPipeline) {

	var expectedScripts *Scripts = NewExecutionPipeline()

	for _, script := range testData {
		expectedScripts.Append(script)
	}

	var actualScripts *Scripts = func() *Scripts {
		o := f()
		switch v := o.(type) {
		case *Scripts:
			return v
		default:
			panic("unable to conver the interface type to concerete type")
		}
	}()
	assert.DeepEqual(t, actualScripts.TestAllScripts(), expectedScripts.TestAllScripts())
	assert.Equal(t, actualScripts.TestLen(), expectedScripts.TestLen())
}
