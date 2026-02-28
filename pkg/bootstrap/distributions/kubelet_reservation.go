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

// KubeletReservation holds computed GKE-style tiered resource reservations.
type KubeletReservation struct {
	KubeReservedCPU      int // millicores
	KubeReservedMemory   int // MiB
	SystemReservedCPU    int // millicores (fixed: 100)
	SystemReservedMemory int // MiB (fixed: 200)
}

// ComputeKubeletReservation returns GKE-style tiered reservations
// given the node's vCPU count and total memory in MiB.
func ComputeKubeletReservation(vcpus, memoryMiB int) KubeletReservation {
	return KubeletReservation{
		KubeReservedCPU:      cpuReserved(vcpus),
		KubeReservedMemory:   memReserved(memoryMiB),
		SystemReservedCPU:    100,
		SystemReservedMemory: 200,
	}
}

func cpuReserved(cpus int) int {
	milli := 0
	if cpus >= 1 {
		milli += 60
	}
	if cpus >= 2 {
		milli += 10
	}
	if cpus >= 3 {
		extra := min(cpus-2, 2)
		milli += extra * 5
	}
	if cpus >= 5 {
		above := cpus - 4
		milli += above * 25 / 10
	}
	return milli
}

func memReserved(mem int) int {
	reserved := 0
	if mem <= 4096 {
		reserved = mem * 25 / 100
	} else if mem <= 8192 {
		reserved = 4096*25/100 + (mem-4096)*20/100
	} else if mem <= 16384 {
		reserved = 4096*25/100 + 4096*20/100 + (mem-8192)*10/100
	} else if mem <= 131072 {
		reserved = 4096*25/100 + 4096*20/100 + 8192*10/100 + (mem-16384)*6/100
	} else {
		reserved = 4096*25/100 + 4096*20/100 + 8192*10/100 + 114688*6/100 + (mem-131072)*2/100
	}
	if reserved < 255 {
		reserved = 255
	}
	return reserved
}

// KubeletReservationScript is a bash snippet that computes GKE-style tiered
// kubelet kube-reserved values at bootstrap time using the node's actual
// CPU count (nproc) and total memory (/proc/meminfo).
//
// It sets two shell variables:
//   - KUBE_CPU  – millicores to reserve (e.g. 60, 70, 80)
//   - KUBE_MEM  – mebibytes to reserve (minimum 255)
//
// The snippet is already %%-escaped so it can be embedded directly inside
// fmt.Sprintf format strings.
const KubeletReservationScript = `
TOTAL_CPUS=$(nproc)
TOTAL_MEM_MI=$(awk '/MemTotal/{printf "%%d", $2/1024}' /proc/meminfo)

cpu_reserved() {
  local cpus=$1
  local milli=0
  if [ "$cpus" -ge 1 ]; then milli=$((milli + 60)); fi
  if [ "$cpus" -ge 2 ]; then milli=$((milli + 10)); fi
  if [ "$cpus" -ge 3 ]; then
    local extra=$((cpus - 2))
    if [ "$extra" -gt 2 ]; then extra=2; fi
    milli=$((milli + extra * 5))
  fi
  if [ "$cpus" -ge 5 ]; then
    local above=$((cpus - 4))
    milli=$((milli + above * 25 / 10))
  fi
  echo "$milli"
}

mem_reserved() {
  local mem=$1
  local reserved=0
  if [ "$mem" -le 4096 ]; then
    reserved=$((mem * 25 / 100))
  elif [ "$mem" -le 8192 ]; then
    reserved=$((4096 * 25 / 100 + (mem - 4096) * 20 / 100))
  elif [ "$mem" -le 16384 ]; then
    reserved=$((4096 * 25 / 100 + 4096 * 20 / 100 + (mem - 8192) * 10 / 100))
  elif [ "$mem" -le 131072 ]; then
    reserved=$((4096 * 25 / 100 + 4096 * 20 / 100 + 8192 * 10 / 100 + (mem - 16384) * 6 / 100))
  else
    reserved=$((4096 * 25 / 100 + 4096 * 20 / 100 + 8192 * 10 / 100 + 114688 * 6 / 100 + (mem - 131072) * 2 / 100))
  fi
  if [ "$reserved" -lt 255 ]; then reserved=255; fi
  echo "$reserved"
}

KUBE_CPU=$(cpu_reserved "$TOTAL_CPUS")
KUBE_MEM=$(mem_reserved "$TOTAL_MEM_MI")
`
