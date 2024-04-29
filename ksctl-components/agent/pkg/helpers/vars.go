package helpers

import (
	"io"
	"os"
)

var (
	LogVerbosity = map[string]int{
		"DEBUG": -1,
		"":      0,
	}

	LogWriter io.Writer = os.Stdout
)
