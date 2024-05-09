package logger

import (
	"strings"

	"github.com/ksctl/ksctl/pkg/helpers/utilities"
)

func addLineTerminationForLongStrings(str string) string {

	//arr with endline split
	arrStr := strings.Split(str, "\n")

	var helper func(string) string

	helper = func(_str string) string {

		if len(_str) < LimitCol {
			return _str
		}

		x := string(utilities.DeepCopySlice[byte]([]byte(_str[:LimitCol])))
		y := string(utilities.DeepCopySlice[byte]([]byte(helper(_str[LimitCol:]))))

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
