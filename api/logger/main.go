package logger

type Logger struct {
	Verbose bool
}

type LogFactory interface {
	// will accept a string to be highlignted
	Info(string, string)
	Warn(string)
	Print(string)
	Err(string)
}
