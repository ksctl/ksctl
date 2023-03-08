package logger

import (
	"fmt"
	"log"
)

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

func (logger *Logger) Info(message, resource string) {

	if logger.Verbose {
		if len(resource) == 0 {
			log.Printf("%s[INFO] %s%v%s", GREEN, WHITE_BOLD, message, RESET)
		} else {
			log.Printf("%s[INFO] %s%v %s%v%s", GREEN, WHITE, message, GREEN, resource, RESET)
		}
	} else {
		if len(resource) == 0 {
			fmt.Printf("%s[INFO] %s%v%s\n", GREEN, WHITE_BOLD, message, RESET)
		} else {
			fmt.Printf("%s[INFO] %s%v %s%v%s\n", GREEN, WHITE, message, GREEN, resource, RESET)
		}
	}
}

func (logger *Logger) Print(message string) {
	if logger.Verbose {
		log.Println("[MSG] ", message)
	} else {
		fmt.Println("[MSG] ", message)
	}
}

func (logger *Logger) Note(message string) {

	if logger.Verbose {
		log.Printf("%s[NOTE] %v%s", BLUE_BOLD, message, RESET)
	} else {
		fmt.Printf("%s[NOTE] %v%s\n", BLUE_BOLD, message, RESET)
	}
}

func (logger *Logger) Warn(message string) {
	if logger.Verbose {
		log.Printf("%s[WARN] %v%s", YELLOW, message, RESET)
	} else {
		fmt.Printf("%s[WARN] %v%s\n", YELLOW, message, RESET)
	}
}

func (logger *Logger) Err(message string) {
	if logger.Verbose {
		log.Printf("%s[ERR] %v%s", RED, message, RESET)
	} else {
		fmt.Printf("%s[ERR] %v%s\n", RED, message, RESET)
	}
}
