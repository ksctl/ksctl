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

package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type IntegrationTestP struct {
	goTag   string
	goBench string
}

func getPackagesIntegrationTest() []IntegrationTestP {
	return []IntegrationTestP{
		{"testing_civo", "BenchmarkCivoTestingHA"},
		{"testing_azure", "BenchmarkAzureTestingHA"},
		{"testing_aws", "BenchmarkAwsTestingHA"},
		{"testing_civo", "BenchmarkCivoTestingManaged"},
		{"testing_azure", "BenchmarkAzureTestingManaged"},
		{"testing_aws", "BenchmarkAwsTestingManaged"},
		{"testing_local", "BenchmarkLocalTestingManaged"},
	}
}

func runTestsIntegration(packages []IntegrationTestP) bool {

	if err := os.RemoveAll(filepath.Join(os.TempDir(), "ksctl-black-box-test")); err != nil {
		fmt.Printf("%s✘ Error cleaning up temporary directory: %s%s\n", RED, err.Error(), RESET)
		return false
	}

	for _, pkg := range packages {
		var cmd *exec.Cmd
		fmt.Printf("\n%sRunning tests for package: %s with tag: %s and bench: %s%s\n", CYAN, pkg, pkg.goTag, pkg.goBench, RESET)

		cmd = exec.Command("go", "test", "-tags", pkg.goTag, "-bench", pkg.goBench, "-benchtime=1x", "-cover", "-v", ".")
		cmd.Dir = "../tests/integration"
		_bout := new(strings.Builder)
		_berr := new(strings.Builder)
		spinner := NewSpinner()

		cmd.Stdout = _bout
		cmd.Stderr = _berr

		spinner.Start()
		err := cmd.Run()
		spinner.Stop()
		if err != nil {
			fmt.Printf("%s✘ Tests failed for package: %+v%s\n", RED, pkg, RESET)
			return false
		}
		fmt.Printf("%s✔ Tests passed for package: %+v%s\n", GREEN, pkg, RESET)
	}
	return true
}

func IntegrationTest(t *testing.T) {
	t.Logf("\n%s%s🧪 Running integration tests...%s\n", CYAN, BOLD, RESET)

	packages := getPackagesIntegrationTest()
	if len(packages) == 0 {
		t.Errorf("%sNo packages to test.%s\n", RED, RESET)
		return
	}
	fmt.Printf("%sFound %d packages to test.%s\n", GREEN, len(packages), RESET)

	allTestsPassed := runTestsIntegration(packages)
	if !allTestsPassed {
		t.Errorf("%s✘ Some tests failed.%s\n", RED, RESET)
		return
	}

	t.Logf("\n%s%s✔ All tests passed.%s\n", GREEN, BOLD, RESET)
}
