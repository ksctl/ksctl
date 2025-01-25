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
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/ksctl/ksctl/hacks"
)

func getPackagesUnitTest() []string {
	EXCLUDE_DIRS := []string{"cmd", "cli", "migration", "vendor", "tests"}

	fmt.Printf("%s%sCollecting packages...%s\n", BLUE, BOLD, RESET)
	cmd := exec.Command("go", "list", "./...")
	cmd.Dir = "../"
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("%sError listing packages: %s%s\n", RED, err.Error(), RESET)
		os.Exit(1)
	}

	packages := strings.Split(string(output), "\n")
	var filteredPackages []string
	for _, pkg := range packages {
		exclude := false
		for _, dir := range EXCLUDE_DIRS {
			if strings.Contains(pkg, dir) {
				exclude = true
				break
			}
		}
		if !exclude && len(pkg) != 0 {
			filteredPackages = append(filteredPackages, pkg)
		}
	}

	fmt.Printf("%sFound %d packages to test.%s\n", GREEN, len(filteredPackages), RESET)
	return filteredPackages
}

func isCloudProviderPackage(pkg string) string {
	re := regexp.MustCompile(`.*/pkg/provider/([^/]+)$`)
	match := re.FindStringSubmatch(pkg)
	if match != nil {
		return match[1]
	}
	return ""
}

func isFuzzPackage(pkg string) bool {
	re := regexp.MustCompile(`.*/pkg/validation$`)
	match := re.FindStringSubmatch(pkg)
	return match != nil
}

func runFuzzTests(packagex string, fuzz string) bool {
	fmt.Printf("\n%sRunning tests for package: %s with Fuzz: %s%s\n", CYAN, packagex, fuzz, RESET)
	cmd := exec.Command("go", "test", "-fuzz", fuzz, "-fuzztime", "10s", "-v", packagex)
	_bout := new(strings.Builder)
	_berr := new(strings.Builder)
	spinner := hacks.NewSpinner()

	cmd.Stdout = _bout
	cmd.Stderr = _berr
	cmd.Dir = "../"

	spinner.Start()
	err := cmd.Run()
	spinner.Stop()
	if err != nil {
		fmt.Printf("%s✘ Tests failed for package: %s%s\n", RED, packagex, RESET)
		fmt.Printf("%s\n%s\n%s\n%s\n", RED, _bout.String(), _berr.String(), RESET)
		return false
	}
	fmt.Printf("%s✔ Tests passed for package: %s%s\n", GREEN, packagex, RESET)
	return true
}

func runTestsUnit(packages []string) bool {
	for _, pkg := range packages {
		var cmd *exec.Cmd
		cloudProvider := isCloudProviderPackage(pkg)
		if cloudProvider != "" {
			goTag := fmt.Sprintf("testing_%s", strings.ToLower(cloudProvider))
			fmt.Printf("\n%sRunning tests for package: %s with tag: %s%s\n", CYAN, pkg, goTag, RESET)
			cmd = exec.Command("go", "test", "-v", "-tags", goTag, pkg)
		} else if isFuzzPackage(pkg) {

			fuzzz := [...]string{"FuzzValidateCloud", "FuzzValidateDistro", "FuzzName", "FuzzValidateRole", "FuzzValidateStorage"}
			for _, fuzz := range fuzzz {
				if !runFuzzTests(pkg, fuzz) {
					fmt.Printf("%s✘ Tests failed for package: %s%s\n", RED, pkg, RESET)
					return false
				}
			}
			fmt.Printf("%s✔ Tests passed for package: %s%s\n", GREEN, pkg, RESET)
			continue
		} else {
			fmt.Printf("\n%sRunning tests for package: %s%s\n", CYAN, pkg, RESET)
			cmd = exec.Command("go", "test", "-fuzz", "Fuzz", "-fuzztime", "10s", "-v", pkg)
		}
		_bout := new(strings.Builder)
		_berr := new(strings.Builder)
		spinner := hacks.NewSpinner()

		cmd.Stdout = _bout
		cmd.Stderr = _berr
		cmd.Dir = "../"

		spinner.Start()
		err := cmd.Run()
		spinner.Stop()

		if err != nil {
			fmt.Printf("%s✘ Tests failed for package: %s%s\n", RED, pkg, RESET)
			fmt.Printf("%s\n%s\n%s\n%s\n", RED, _bout.String(), _berr.String(), RESET)
			return false
		}
		fmt.Printf("%s✔ Tests passed for package: %s%s\n", GREEN, pkg, RESET)
	}
	return true
}

func UnitTest(t *testing.T) {
	t.Logf("\n%s%s🧪 Running unit tests...%s\n", CYAN, BOLD, RESET)

	packages := getPackagesUnitTest()
	if len(packages) == 0 {
		t.Errorf("%sNo packages to test.%s\n", RED, RESET)
		return
	}

	allTestsPassed := runTestsUnit(packages)

	if !allTestsPassed {
		t.Errorf("\n%s%s🚨 Some tests failed. Check the output above for details.%s\n", RED, BOLD, RESET)
		return
	}

	t.Logf("\n%s%s🎉 All tests passed successfully!%s\n", GREEN, BOLD, RESET)
}
