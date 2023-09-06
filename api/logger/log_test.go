package logger

import (
	"os"
	"testing"
)

var (
	logger LogFactory
)

func TestMain(m *testing.M) {
	logger = &Logger{Verbose: true}
	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestPrinters(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		logger.Success("FAKE")
	})

	t.Run("Warn", func(t *testing.T) {
		logger.Warn("FAKE")
	})

	t.Run("Err", func(t *testing.T) {
		logger.Err("FAKE")
	})

	t.Run("Note", func(t *testing.T) {
		logger.Note("FAKE")
	})

	t.Run("Print", func(t *testing.T) {
		logger.Print("FAKE")
	})

	t.Run("Table", func(t *testing.T) {
		logger.Table(nil)
	})
}
