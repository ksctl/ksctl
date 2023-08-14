package logger

import (
	"fmt"
	"github.com/fatih/color"
	"log"
	"strings"
)

// Info information message to stdout
// in green colour
func (logger *Logger) Success(message ...string) {
	color.Set(color.FgGreen, color.Bold)
	defer color.Unset()
	outputMsg := strings.Join(message, " ")

	if logger.Verbose {
		log.Printf("[SUCCESS] %v", outputMsg)
	} else {
		fmt.Printf("[SUCCESS] %v\n", outputMsg)
	}
}

// Print plan text stdout
func (logger *Logger) Print(message ...string) {
	outputMsg := strings.Join(message, " ")
	if logger.Verbose {
		log.Println("[LOG]", outputMsg)
	} else {
		fmt.Println("[LOG]", outputMsg)
	}
}

// Note note taking message to stdout
// in blue colour
func (logger *Logger) Note(message ...string) {

	color.Set(color.FgBlue, color.Bold)
	defer color.Unset()

	outputMsg := strings.Join(message, " ")
	if logger.Verbose {
		log.Printf("[NOTE] %v", outputMsg)
	} else {
		fmt.Printf("[NOTE] %v\n", outputMsg)
	}
}

// Warn warning message to stdout
// in yellow colour
func (logger *Logger) Warn(message ...string) {
	color.Set(color.FgYellow, color.Bold)
	defer color.Unset()

	outputMsg := strings.Join(message, " ")
	if logger.Verbose {
		log.Printf("[WARN] %v", outputMsg)
	} else {
		fmt.Printf("[WARN] %v\n", outputMsg)
	}
}

// Err error message to stdout
// in red color
func (logger *Logger) Err(message ...string) {
	color.Set(color.FgHiRed, color.Bold)
	defer color.Unset()

	outputMsg := strings.Join(message, " ")
	if logger.Verbose {
		log.Printf("[ERR] %v", outputMsg)
	} else {
		fmt.Printf("[ERR] %v\n", outputMsg)
	}
}
