package logger

type Logger struct {
	Msg string
}

type LogFactory interface {
	// will accept a string to be highlignted
	Info(string)
	Warn()
	Err()
}
