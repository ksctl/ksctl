package helpers

import (
	"context"
	"github.com/ksctl/ksctl/pkg/logger"
	"os"
	"regexp"
	"testing"
)

func FuzzName(f *testing.F) {
	testcases := []string{"avcd", "nice-23rde-fe423"}

	for _, tc := range testcases {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}

	f.Fuzz(func(t *testing.T, name string) {
		outErr := IsValidName(context.TODO(), logger.NewStructuredLogger(-1, os.Stdout), name)
		t.Logf("name: %s and err: %v", name, outErr)
		matched, err := regexp.MatchString(`(^[a-z])([-a-z0-9])*([a-z0-9]$)`, name)

		if outErr == nil && len(name) > 50 {
			t.Errorf("incorrect error for long length string")
		}
		if outErr != nil && (!matched || err != nil) {
			return
		}
		if outErr == nil && (matched && err == nil) {
			return
		}
	})
}
