package logger

import (
	"fmt"
	"os"
	"strings"
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

func TestHelperToAddLineTerminationForLongStrings(t *testing.T) {
	test := fmt.Sprintf("Argo Rollouts (Ver: %s) is a Kubernetes controller and set of CRDs which provide advanced deployment capabilities such as blue-green, canary, canary analysis, experimentation, and progressive delivery features to Kubernetes.", "v1.2.4")

	x := strings.Split(addLineTerminationForLongStrings(test), "\n")
	for _, line := range x {
		if len(line) > LimitCol+1 {
			t.Errorf("Line too long: %s, got: %d, expected: %d", line, len(line), LimitCol)
		}
	}
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
