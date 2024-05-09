package logger

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
	"github.com/ksctl/ksctl/pkg/resources/controllers/cloud"
)

var (
	sL resources.LoggerFactory
	gL resources.LoggerFactory
)

func TestMain(m *testing.M) {
	sL = NewStructuredLogger(-1, os.Stdout)
	_ = NewStructuredLogger(0, os.Stdout)

	gL = NewGeneralLogger(-1, os.Stdout)
	_ = NewGeneralLogger(0, os.Stdout)
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

func TestPrintersStructured(t *testing.T) {

	t.Run("Package name set", func(t *testing.T) {
		sL.SetPackageName("logger-test")
	})

	t.Run("Success", func(t *testing.T) {
		sL.Success("FAKE", "type", "success")
	})

	t.Run("Warn", func(t *testing.T) {
		sL.Warn("FAKE", "type", "warn")
	})

	t.Run("Error", func(t *testing.T) {
		sL.Error("FAKE", "type", "error")
	})

	t.Run("Debug", func(t *testing.T) {
		sL.Debug("FAKE", "type", "debugging")
	})

	t.Run("Note", func(t *testing.T) {
		sL.Note("FAKE", "type", "note")
	})

	t.Run("Print", func(t *testing.T) {
		sL.Print("FAKE", "type", "print")
	})

	t.Run("Table", func(t *testing.T) {
		sL.Table(
			[]cloud.AllClusterData{
				{
					Name:     "fake-demo",
					Provider: "fake",
					Region:   "fake-reg",
				},
			})

		sL.Table(nil)
	})

	t.Run("Box", func(t *testing.T) {
		sL.Box("Abcd", "1")
		sL.Box("Abcddedefe", "1")
		sL.Box("KUBECONFIG env var", "/jknc/csdc")
		sL.Box("KUBECONFIG env var", "jknc")
	})

	t.Run("external", func(t *testing.T) {
		sL.ExternalLogHandler(consts.LOG_SUCCESS, "cdcc")
		sL.ExternalLogHandlerf(consts.LOG_SUCCESS, "cdcc: %v", nil)
	})
}

func TestPrintersGeneral(t *testing.T) {
	t.Run("Package name set", func(t *testing.T) {
		gL.SetPackageName("logger-test")
	})

	t.Run("Success", func(t *testing.T) {
		gL.Success("FAKE", "type", "success")
	})

	t.Run("Warn", func(t *testing.T) {
		gL.Warn("FAKE", "type", "warn")
	})

	t.Run("Error", func(t *testing.T) {
		gL.Error("FAKE", "type", "error")
	})

	t.Run("Debug", func(t *testing.T) {
		gL.Debug("FAKE", "type", "debugging")
	})

	t.Run("Note", func(t *testing.T) {
		gL.Note("FAKE", "type", "note")
	})

	t.Run("Print", func(t *testing.T) {
		gL.Print("FAKE", "type", "print")
	})

	t.Run("Table", func(t *testing.T) {
		gL.Table(
			[]cloud.AllClusterData{
				{
					Name:     "fake-demo",
					Provider: "fake",
					Region:   "fake-reg",
				},
			})

		gL.Table(nil)
	})

	t.Run("Box", func(t *testing.T) {
		gL.Box("Abcd", "1")
		gL.Box("Abcddedefe", "1")
		gL.Box("KUBECONFIG env var", "/jknc/csdc")
		gL.Box("KUBECONFIG env var", "jknc")
	})

	t.Run("external", func(t *testing.T) {
		gL.ExternalLogHandler(consts.LOG_SUCCESS, "cdcc")
		gL.ExternalLogHandlerf(consts.LOG_SUCCESS, "cdcc", "Reason", fmt.Errorf("Error"))
	})
}
