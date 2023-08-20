package test

import (
	"github.com/kubesimplify/ksctl/api/utils"
	"os"
	"testing"
)

func BenchmarkCivoTestingManaged(b *testing.B) {
	if err := os.Setenv(utils.FAKE_CLIENT, "1"); err != nil {
		b.Fatalf("Failed to set fake env %v", err)
	}
	StartCloud()

	for i := 0; i < b.N; i++ {
		if err := CivoTestingManaged(); err != nil {
			b.Fatalf("failed, err: %v", err)
		}
	}
}

func BenchmarkCivoTestingHA(b *testing.B) {
	if err := os.Setenv(utils.FAKE_CLIENT, "1"); err != nil {
		b.Fatalf("Failed to set fake env %v", err)
	}
	StartCloud()

	for i := 0; i < b.N; i++ {
		if err := CivoTestingHA(); err != nil {
			b.Fatalf("failed, err: %v", err)
		}
	}
}
