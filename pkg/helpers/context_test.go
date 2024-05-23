package helpers

import (
	"context"
	"fmt"
	"testing"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"gotest.tools/v3/assert"
)

func TestIsContextPresent(t *testing.T) {
	type cases struct {
		ctx      context.Context
		key      consts.KsctlContextKeyType
		expected bool
	}

	ppCtx := context.Background()
	testCases := [...]cases{
		{
			ctx:      context.WithValue(ppCtx, consts.KsctlTestFlagKey, "true"),
			key:      consts.KsctlTestFlagKey,
			expected: true,
		},
		{
			ctx:      context.WithValue(ppCtx, consts.KsctlModuleNameKey, "abcd"),
			key:      consts.KsctlModuleNameKey,
			expected: true,
		},
		{
			ctx:      context.WithValue(ppCtx, consts.KsctlCustomDirLoc, "/tmp/abcd sfvvs"),
			key:      consts.KsctlCustomDirLoc,
			expected: true,
		},
		{
			ctx:      context.WithValue(ppCtx, consts.KsctlTestFlagKey, ""),
			key:      consts.KsctlTestFlagKey,
			expected: false,
		},
	}

	for _, tt := range testCases {
		t.Run(fmt.Sprintf("test case on, %#v", tt.ctx), func(t *testing.T) {
			v, got := IsContextPresent(tt.ctx, tt.key)
			assert.Equal(t, got, tt.expected)
			assert.Equal(t, v, tt.ctx.Value(tt.key))
		})
	}
}
