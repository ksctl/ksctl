package logger

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	"github.com/ksctl/ksctl/pkg/types/controllers/cloud"
)

var (
	sL       types.LoggerFactory
	dummyCtx = context.TODO()
)

func TestMain(m *testing.M) {
	sL = NewStructuredLogger(-1, os.Stdout)
	_ = NewStructuredLogger(0, os.Stdout)

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
		sL.Error("FAKE", "type", "error")
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
		sL.Table(dummyCtx, consts.LoggingGetClusters,
			[]cloud.AllClusterData{
				{
					Name:          "fake-demo",
					CloudProvider: "fake",
					Region:        "fake-reg",
				},
			})

		sL.Table(dummyCtx, consts.LoggingInfoCluster, nil)
	})

	t.Run("Box", func(t *testing.T) {
		sL.Box(dummyCtx, "Abcd", "1")
		sL.Box(dummyCtx, "Abcddedefe", "1")
		sL.Box(dummyCtx, "KUBECONFIG env var", "/jknc/csdc")
		sL.Box(dummyCtx, "KUBECONFIG env var", "jknc")
	})

	t.Run("external", func(t *testing.T) {
		for _, logType := range []consts.CustomExternalLogLevel{
			consts.LogSuccess, consts.LogError, consts.LogWarning, consts.LogNote,
		} {
			sL.ExternalLogHandler(dummyCtx, logType, "cdcc")
			sL.ExternalLogHandlerf(dummyCtx, logType, "cdcc: %v", nil)
		}
	})
}

func TestFormGroupsHandlesEmptyInput(t *testing.T) {
	format, vals := formGroups()
	assert.Equal(t, "", format)
	assert.Nil(t, vals)
}

func TestFormGroupsHandlesSingleKeyValuePair(t *testing.T) {
	format, vals := formGroups("key", "value")
	assert.Equal(t, "key=%v", format)
	assert.Equal(t, []any{"value"}, vals)
}

func TestFormGroupsHandlesMultipleKeyValuePairs(t *testing.T) {
	format, vals := formGroups("key1", "value1", "key2", "value2")
	assert.Equal(t, "key1=%v key2=%v", format)
	assert.Equal(t, []any{"value1", "value2"}, vals)
}

func TestFormGroupsHandlesExtraValues(t *testing.T) {
	format, vals := formGroups("key1", "value1", "key2")
	assert.Equal(t, "key1=%v !!EXTRA:%v", format)
	assert.Equal(t, []any{"value1", "key2"}, vals)
}

func TestFormGroupsHandlesStructValues(t *testing.T) {
	type TestStruct struct {
		Field string
	}
	format, vals := formGroups("key", TestStruct{Field: "value"})
	assert.Equal(t, "key=%#v", format)
	assert.Equal(t, []any{TestStruct{Field: "value"}}, vals)
}

func TestFormGroupsHandlesErrorValues(t *testing.T) {
	err := fmt.Errorf("test error")
	format, vals := formGroups("key", err)
	assert.Equal(t, "key=%v", format)
	assert.Equal(t, []any{err}, vals)
}
