package stacks

import (
	"context"
	"os"
	"testing"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestFetchKsctlStackWithValidStackID(t *testing.T) {
	ctx := context.Background()
	log := logger.NewStructuredLogger(-1, os.Stdout)
	fn, err := FetchKsctlStack(ctx, log, string(metadata.ArgocdStandardStackID))
	assert.NoError(t, err)
	assert.NotNil(t, fn)
}

func TestFetchKsctlStackWithInvalidStackID(t *testing.T) {
	ctx := context.Background()
	log := logger.NewStructuredLogger(-1, os.Stdout)
	fn, err := FetchKsctlStack(ctx, log, "invalidStackID")
	assert.Error(t, err)
	assert.Nil(t, fn)
}

func TestFetchKsctlStackWithEmptyStackID(t *testing.T) {
	ctx := context.Background()
	log := logger.NewStructuredLogger(-1, os.Stdout)
	fn, err := FetchKsctlStack(ctx, log, "")
	assert.Error(t, err)
	assert.Nil(t, fn)
}
