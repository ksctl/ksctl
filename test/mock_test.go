package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/ksctl/ksctl/pkg/helpers"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

func BenchmarkCivoTestingManaged(b *testing.B) {
	if err := os.Setenv(string(consts.KsctlFakeFlag), "1"); err != nil {
		b.Fatalf("Failed to set fake env %v", err)
	}
	if err := InitCore(); err != nil {
		b.Fatalf("failed to start core: %v", err)
	}

	for i := 0; i < b.N; i++ {
		if err := CivoTestingManaged(); err != nil {
			b.Fatalf("failed, err: %v", err)
		}
	}

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(os.TempDir() + helpers.PathSeparator + "ksctl-black-box-test"); err != nil {
		panic(err)
	}
}

func BenchmarkCivoTestingHA(b *testing.B) {
	if err := os.Setenv(string(consts.KsctlFakeFlag), "1"); err != nil {
		b.Fatalf("Failed to set fake env %v", err)
	}
	if err := InitCore(); err != nil {
		b.Fatalf("failed to start core: %v", err)
	}

	for i := 0; i < b.N; i++ {
		if err := CivoTestingHAK3s(); err != nil {
			b.Fatalf("failed, err: %v", err)
		}
		if err := CivoTestingHAKubeadm(); err != nil {
			b.Fatalf("failed, err: %v", err)
		}
	}

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(os.TempDir() + helpers.PathSeparator + "ksctl-black-box-test"); err != nil {
		panic(err)
	}
}

func BenchmarkAzureTestingHA(b *testing.B) {
	if err := os.Setenv(string(consts.KsctlFakeFlag), "1"); err != nil {
		b.Fatalf("Failed to set fake env %v", err)
	}
	if err := InitCore(); err != nil {
		b.Fatalf("failed to start core: %v", err)
	}

	for i := 0; i < b.N; i++ {
		if err := AzureTestingHAK3s(); err != nil {
			b.Fatalf("failed, err: %v", err)
		}
		if err := AzureTestingHAKubeadm(); err != nil {
			b.Fatalf("failed, err: %v", err)
		}
	}

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(os.TempDir() + helpers.PathSeparator + "ksctl-black-box-test"); err != nil {
		panic(err)
	}
}

func BenchmarkAwsTestingHA(b *testing.B) {
	if err := os.Setenv(string(consts.KsctlFakeFlag), "1"); err != nil {
		b.Fatalf("Failed to set fake env %v", err)
	}
	if err := InitCore(); err != nil {
		b.Fatalf("failed to start core: %v", err)
	}

	for i := 0; i < b.N; i++ {
		if err := AwsTestingHA(); err != nil {
			b.Fatalf("failed, err: %v", err)
		}
	}

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(os.TempDir() + helpers.PathSeparator + "ksctl-black-box-test"); err != nil {
		panic(err)
	}
}

func BenchmarkAzureTestingManaged(b *testing.B) {
	if err := os.Setenv(string(consts.KsctlFakeFlag), "1"); err != nil {
		b.Fatalf("Failed to set fake env %v", err)
	}
	if err := InitCore(); err != nil {
		b.Fatalf("failed to start core: %v", err)
	}

	for i := 0; i < b.N; i++ {
		if err := AzureTestingManaged(); err != nil {
			b.Fatalf("failed, err: %v", err)
		}
	}

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(os.TempDir() + helpers.PathSeparator + "ksctl-black-box-test"); err != nil {
		panic(err)
	}
}

func BenchmarkLocalTestingManaged(b *testing.B) {
	if err := os.Setenv(string(consts.KsctlFakeFlag), "1"); err != nil {
		b.Fatalf("Failed to set fake env %v", err)
	}
	if err := InitCore(); err != nil {
		b.Fatalf("failed to start core: %v", err)
	}

	for i := 0; i < b.N; i++ {
		if err := LocalTestingManaged(); err != nil {
			b.Fatalf("failed, err: %v", err)
		}
	}

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(os.TempDir() + helpers.PathSeparator + "ksctl-black-box-test"); err != nil {
		panic(err)
	}
}
