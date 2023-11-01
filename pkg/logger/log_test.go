package logger

import (
	"os"
	"testing"

	"github.com/kubesimplify/ksctl/pkg/resources"
)

var (
	logger resources.LoggerFactory
)

func TestMain(m *testing.M) {
	logger = NewDefaultLogger(-1, os.Stdout)
	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestPrinters(t *testing.T) {

	t.Run("Paclage name set", func(t *testing.T) {
		logger.SetPackageName("logger-test")
	})

	t.Run("Success", func(t *testing.T) {
		logger.Success("FAKE")
	})

	t.Run("Warn", func(t *testing.T) {
		logger.Warn("FAKE")
	})

	t.Run("Err", func(t *testing.T) {
		logger.Error("FAKE")
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

	t.Run("Box", func(t *testing.T) {
		logger.Box("Abcd", "1")
		logger.Box("Abcddedefe", "1")
		logger.Box("KUBECONFIG env var", "/jknc/csdc")
		logger.Box("KUBECONFIG env var", "jknc")
	})
}
