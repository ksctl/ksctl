package test

import (
	"fmt"
	"os"
	"testing"
)

func BenchmarkCivoTestingManaged(b *testing.B) {
	if err := InitCore(); err != nil {
		b.Fatalf("failed to start core: %v", err)
	}

	for i := 0; i < b.N; i++ {
		if err := CivoTestingManaged(); err != nil {
			b.Fatalf("failed, err: %v", err)
		}
	}

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}
}

func BenchmarkCivoTestingHA(b *testing.B) {
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
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}
}

func BenchmarkAzureTestingHA(b *testing.B) {
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
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}
}

func BenchmarkAwsTestingHA(b *testing.B) {
	if err := InitCore(); err != nil {
		b.Fatalf("failed to start core: %v", err)
	}

	for i := 0; i < b.N; i++ {
		if err := AwsTestingHA(); err != nil {
			b.Fatalf("failed, err: %v", err)
		}
	}

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}
}

func BenchmarkAzureTestingManaged(b *testing.B) {
	if err := InitCore(); err != nil {
		b.Fatalf("failed to start core: %v", err)
	}

	for i := 0; i < b.N; i++ {
		if err := AzureTestingManaged(); err != nil {
			b.Fatalf("failed, err: %v", err)
		}
	}

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}
}

func BenchmarkLocalTestingManaged(b *testing.B) {
	if err := InitCore(); err != nil {
		b.Fatalf("failed to start core: %v", err)
	}

	for i := 0; i < b.N; i++ {
		if err := LocalTestingManaged(); err != nil {
			b.Fatalf("failed, err: %v", err)
		}
	}

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}
}
