package logger

import "log"

const (
	GREEN      = "\033[1;32m"
	RED        = "\033[1;31m"
	YELLOW     = "\033[1;33m"
	BLUE       = "\033[0;34m"
	BLUE_BOLD  = "\033[1;34m"
	WHITE      = "\033[0;0m"
	WHITE_BOLD = "\033[1;1m"
	RESET      = "\033[0m"
)

func (logger *Logger) Info(resource string) {
	if len(resource) == 0 {
		log.Printf("%s%v%s", WHITE_BOLD, logger.Msg, RESET)
	} else {
		log.Printf("%s%v %s%v%s", WHITE, logger.Msg, GREEN, resource, RESET)
	}
}

func (logger *Logger) Warn() {
	log.Printf("%s%v%s", YELLOW, logger.Msg, RESET)
}

func (Logger *Logger) Err() {
	log.Printf("%s%v%s", RED, Logger.Msg, RESET)
}
