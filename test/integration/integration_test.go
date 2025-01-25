// Copyright 2024 Ksctl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration

import (
	"fmt"
	"os"
	"testing"
)

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

func BenchmarkAwsTestingManaged(b *testing.B) {
	if err := InitCore(); err != nil {
		b.Fatalf("failed to start core: %v", err)
	}

	for i := 0; i < b.N; i++ {
		if err := AwsTestingManaged(); err != nil {
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
