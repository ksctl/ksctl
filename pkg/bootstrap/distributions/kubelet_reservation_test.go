// Copyright 2026 Ksctl Authors
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

package distributions

import "testing"

func TestComputeKubeletReservation(t *testing.T) {
	tests := []struct {
		name       string
		vcpus      int
		memoryMiB  int
		wantCPU    int
		wantMemory int
	}{
		{
			name:       "1 vCPU, 1024 MiB (small node, hits 255 MiB minimum)",
			vcpus:      1,
			memoryMiB:  1024,
			wantCPU:    60,
			wantMemory: 256,
		},
		{
			name:       "2 vCPU, 4096 MiB (common small instance)",
			vcpus:      2,
			memoryMiB:  4096,
			wantCPU:    70,
			wantMemory: 1024,
		},
		{
			name:       "4 vCPU, 16384 MiB (standard instance)",
			vcpus:      4,
			memoryMiB:  16384,
			wantCPU:    80,
			wantMemory: 2662,
		},
		{
			name:       "8 vCPU, 32768 MiB (larger instance)",
			vcpus:      8,
			memoryMiB:  32768,
			wantCPU:    90,
			wantMemory: 3645,
		},
		{
			name:       "16 vCPU, 131072 MiB (large instance)",
			vcpus:      16,
			memoryMiB:  131072,
			wantCPU:    110,
			wantMemory: 9543,
		},
		{
			name:       "96 vCPU, 196608 MiB (very large instance, hits top tier)",
			vcpus:      96,
			memoryMiB:  196608,
			wantCPU:    310,
			wantMemory: 10853,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeKubeletReservation(tt.vcpus, tt.memoryMiB)

			if got.KubeReservedCPU != tt.wantCPU {
				t.Errorf("KubeReservedCPU = %d, want %d", got.KubeReservedCPU, tt.wantCPU)
			}
			if got.KubeReservedMemory != tt.wantMemory {
				t.Errorf("KubeReservedMemory = %d, want %d", got.KubeReservedMemory, tt.wantMemory)
			}
			if got.SystemReservedCPU != 100 {
				t.Errorf("SystemReservedCPU = %d, want 100", got.SystemReservedCPU)
			}
			if got.SystemReservedMemory != 200 {
				t.Errorf("SystemReservedMemory = %d, want 200", got.SystemReservedMemory)
			}
		})
	}
}
