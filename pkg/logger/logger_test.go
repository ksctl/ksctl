package logger

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	"github.com/ksctl/ksctl/pkg/types/controllers/cloud"
)

var (
	sL       types.LoggerFactory
	gL       types.LoggerFactory
	dummyCtx = context.TODO()
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
		if len(line) > limitCol+1 {
			t.Errorf("Line too long: %s, got: %d, expected: %d", line, len(line), limitCol)
		}
	}
}

func TestPrintersStructured(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		sL.Success(dummyCtx, "FAKE", "type", "success")
	})

	t.Run("Warn", func(t *testing.T) {
		sL.Warn(dummyCtx, "FAKE", "type", "warn")
	})

	t.Run("Error", func(t *testing.T) {
		sL.Error(dummyCtx, "FAKE", "type", "error")
	})

	t.Run("Debug", func(t *testing.T) {
		sL.Debug(dummyCtx, "FAKE", "type", "debugging")
	})

	t.Run("Note", func(t *testing.T) {
		sL.Note(dummyCtx, "FAKE", "type", "note")
	})

	t.Run("Print", func(t *testing.T) {
		sL.Print(dummyCtx, "FAKE", "type", "print")
	})

	t.Run("Table", func(t *testing.T) {
		sL.Table(dummyCtx,
			[]cloud.AllClusterData{
				{
					Name:          "fake-demo",
					CloudProvider: "fake",
					Region:        "fake-reg",
				},
			})

		sL.Table(dummyCtx, nil)
	})

	t.Run("Box", func(t *testing.T) {
		sL.Box(dummyCtx, "Abcd", "1")
		sL.Box(dummyCtx, "Abcddedefe", "1")
		sL.Box(dummyCtx, "KUBECONFIG env var", "/jknc/csdc")
		sL.Box(dummyCtx, "KUBECONFIG env var", "jknc")
	})

	t.Run("external", func(t *testing.T) {
		sL.ExternalLogHandler(dummyCtx, consts.LogSuccess, "cdcc")
		sL.ExternalLogHandlerf(dummyCtx, consts.LogSuccess, "cdcc: %v", nil)
	})
}

func TestPrintersGeneral(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		gL.Success(dummyCtx, "FAKE", "type", "success")
	})

	t.Run("Warn", func(t *testing.T) {
		gL.Warn(dummyCtx, "FAKE", "type", "warn")
	})

	t.Run("Error", func(t *testing.T) {
		gL.Error(dummyCtx, "FAKE", "type", "error")
	})

	t.Run("Debug", func(t *testing.T) {
		gL.Debug(dummyCtx, "FAKE", "type", "debugging")
	})

	t.Run("Note", func(t *testing.T) {
		gL.Note(dummyCtx, "FAKE", "type", "note")
	})

	t.Run("Print", func(t *testing.T) {
		gL.Print(dummyCtx, "FAKE", "type", "print")
	})

	t.Run("Table", func(t *testing.T) {
		gL.Table(dummyCtx,
			[]cloud.AllClusterData{
				{
					Name:          "fake-demo",
					CloudProvider: "fake",
					Region:        "fake-reg",
				},
			})

		gL.Table(dummyCtx, nil)
	})

	t.Run("Box", func(t *testing.T) {
		gL.Box(dummyCtx, "Abcd", "1")
		gL.Box(dummyCtx, "Abcddedefe", "1")
		gL.Box(dummyCtx, "KUBECONFIG env var", "/jknc/csdc")
		gL.Box(dummyCtx, "KUBECONFIG env var", "jknc")
	})

	t.Run("external", func(t *testing.T) {
		gL.ExternalLogHandler(dummyCtx, consts.LogSuccess, "cdcc")
		gL.ExternalLogHandlerf(dummyCtx, consts.LogSuccess, "cdcc", "Reason", fmt.Errorf("Error"))
	})
}
