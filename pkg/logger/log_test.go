package logger

import (
	"os"
	"testing"

	"github.com/ksctl/ksctl/pkg/resources"
	"github.com/ksctl/ksctl/pkg/resources/controllers/cloud"
)

var (
	logger resources.LoggerFactory
)

func TestMain(m *testing.M) {
	logger = NewDefaultLogger(-1, os.Stdout)
	_ = NewDefaultLogger(0, os.Stdout)
	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestPrinters(t *testing.T) {

	t.Run("Package name set", func(t *testing.T) {
		logger.SetPackageName("logger-test")
	})

	t.Run("Success", func(t *testing.T) {
		logger.Success("FAKE", "type", "success")
	})

	t.Run("Warn", func(t *testing.T) {
		logger.Warn("FAKE", "type", "warn")
	})

	t.Run("Error", func(t *testing.T) {
		logger.Error("FAKE", "type", "error")
	})

	t.Run("Debug", func(t *testing.T) {
		logger.Debug("FAKE", "type", "debugging")
	})

	t.Run("Note", func(t *testing.T) {
		logger.Note("FAKE", "type", "note")
	})

	t.Run("Print", func(t *testing.T) {
		logger.Print("FAKE", "type", "print")
	})

	t.Run("Table", func(t *testing.T) {
		logger.Table(
			[]cloud.AllClusterData{
				cloud.AllClusterData{
					Name:     "fake-demo",
					Provider: "fake",
					Region:   "fake-reg",
				},
			})

		logger.Table(nil)
	})

	t.Run("Box", func(t *testing.T) {
		logger.Box("Abcd", "1")
		logger.Box("Abcddedefe", "1")
		logger.Box("KUBECONFIG env var", "/jknc/csdc")
		logger.Box("KUBECONFIG env var", "jknc")
	})
}
