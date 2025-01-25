// Copyright 2024 ksctl
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

package certs

import (
	"context"
	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"os"
	"testing"
)

var (
	log logger.Logger = logger.NewStructuredLogger(-1, os.Stdout)
)

func TestGenerateCerts(t *testing.T) {
	if ca, etcd, key, err := GenerateCerts(
		context.WithValue(
			context.TODO(),
			consts.KsctlModuleNameKey,
			"demo"),
		log, []string{"192.168.1.1"}); err != nil {
		t.Fatalf("it shouldn't fail, ca: %v, etcd: %v, key: %v, err: %v\n", ca, etcd, key, err)
	}

	if ca, etcd, key, err := GenerateCerts(
		context.WithValue(context.TODO(), consts.KsctlModuleNameKey, "demo"), log, []string{"192,168.1.1"}); err == nil ||
		len(ca) != 0 ||
		len(etcd) != 0 ||
		len(key) != 0 {
		t.Fatalf("it should fail, ca: %v, etcd: %v, key: %v, err: %v\n", ca, etcd, key, err)
	}
}
