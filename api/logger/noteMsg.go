package logger

import (
	"fmt"
	"log"
	"strings"
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

// Info information message to stdout
// in green colour
func (logger *Logger) Info(message ...string) {

	outputMsg := strings.Join(message, " ")

	if logger.Verbose {
		log.Printf("%s[INFO] %v%s", GREEN, outputMsg, RESET)
	} else {
		fmt.Printf("%s[INFO] %v%s\n", GREEN, outputMsg, RESET)
	}
}

// Print plan text stdout
func (logger *Logger) Print(message ...string) {
	outputMsg := strings.Join(message, " ")
	if logger.Verbose {
		log.Println("[TXT] ", outputMsg)
	} else {
		fmt.Println("[TXT] ", outputMsg)
	}
}

// Note note taking message to stdout
// in blue colour
func (logger *Logger) Note(message ...string) {

	outputMsg := strings.Join(message, " ")
	if logger.Verbose {
		log.Printf("%s[NOTE] %v%s", BLUE_BOLD, outputMsg, RESET)
	} else {
		fmt.Printf("%s[NOTE] %v%s\n", BLUE_BOLD, outputMsg, RESET)
	}
}

// Warn warning message to stdout
// in yellow colour
func (logger *Logger) Warn(message ...string) {
	outputMsg := strings.Join(message, " ")
	if logger.Verbose {
		log.Printf("%s[WARN] %v%s", YELLOW, outputMsg, RESET)
	} else {
		fmt.Printf("%s[WARN] %v%s\n", YELLOW, outputMsg, RESET)
	}
}

// Err error message to stdout
// in red color
func (logger *Logger) Err(message ...string) {
	outputMsg := strings.Join(message, " ")
	if logger.Verbose {
		log.Printf("%s[ERR] %v%s", RED, outputMsg, RESET)
	} else {
		fmt.Printf("%s[ERR] %v%s\n", RED, outputMsg, RESET)
	}
}
