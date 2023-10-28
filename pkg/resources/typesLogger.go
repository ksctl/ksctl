package resources

import (
	"io"
)

type LoggerFactory interface {
	Print(string, ...any)

	Success(string, ...any)

	Note(string, ...any)

	Warn(string, ...any)

	Error(string, ...any)

	Debug(string, ...any)

	New(int, io.Writer)

	SetModule(string)
}
