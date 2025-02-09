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

package logger

import (
	"strings"

	"github.com/fatih/color"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

type CustomExternalLogLevel string
type LogClusterDetail uint

var (
	LogWarning = CustomExternalLogLevel(color.New(color.FgBlack, color.BgYellow).Sprintf("WARN"))
	LogInfo    = CustomExternalLogLevel(color.New(color.FgBlack, color.BgBlue).Sprintf("INFO"))
	LogNote    = CustomExternalLogLevel(color.New(color.FgBlack, color.BgCyan).Sprintf("NOTE"))
	LogDebug   = CustomExternalLogLevel(color.New(color.FgBlack, color.BgMagenta).Sprintf("DEBUG"))
	LogSuccess = CustomExternalLogLevel(color.New(color.FgBlack, color.BgGreen).Sprintf("SUCC"))
	LogError   = CustomExternalLogLevel(color.New(color.FgBlack, color.BgRed).Sprintf("ERR"))
)

const (
	LoggingGetClusters LogClusterDetail = iota
	LoggingInfoCluster LogClusterDetail = iota
)

func addLineTerminationForLongStrings(str string) string {

	//arr with endline split
	arrStr := strings.Split(str, "\n")

	var helper func(string) string

	helper = func(_str string) string {

		if len(_str) < limitCol {
			return _str
		}

		x := string(utilities.DeepCopySlice[byte]([]byte(_str[:limitCol])))
		y := string(utilities.DeepCopySlice[byte]([]byte(helper(_str[limitCol:]))))

		// ks
		// ^^
		if x[len(x)-1] != ' ' && y[0] != ' ' {
			x += "-"
		}

		_new := x + "\n" + y
		return _new
	}

	for idx, line := range arrStr {
		arrStr[idx] = helper(line)
	}

	return strings.Join(arrStr, "\n")
}
