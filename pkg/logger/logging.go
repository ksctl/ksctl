package logger

import (
	"github.com/fatih/color"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"strings"
)

type CustomExternalLogLevel string
type LogClusterDetail uint

var (
	LogWarning = CustomExternalLogLevel(color.HiYellowString("WARN"))
	LogError   = CustomExternalLogLevel(color.HiRedString("ERR"))
	LogInfo    = CustomExternalLogLevel(color.HiBlueString("INFO"))
	LogNote    = CustomExternalLogLevel(color.CyanString("NOTE"))
	LogSuccess = CustomExternalLogLevel(color.HiGreenString("PASS"))
	LogDebug   = CustomExternalLogLevel(color.WhiteString("DEBUG"))
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
