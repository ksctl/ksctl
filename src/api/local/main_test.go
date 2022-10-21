package local

import (
	"strings"
	"testing"
)

func TestGenerateConfig(t *testing.T) {
	var rawConfig []byte
	var testRawConfig string
	var err error

	rawConfig, err = generateConfig(0, 0)

	if err != nil {
		if err.Error() != "invalid config request control node cannot be 0" {
			t.Fatalf("ERR Bad node config was [PASSED]!")
		}
	}

	rawConfig, err = generateConfig(2, 0)

	if err != nil {
		if err.Error() != "invalid config request control node cannot be 0" {
			t.Fatalf("ERR Bad node config was [PASSED]!")
		}
	}
	rawConfig, err = generateConfig(2, 2)
	testRawConfig = `---
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: control-plane
- role: worker
- role: worker
...`
	if strings.Compare(testRawConfig, string(rawConfig)) != 0 {
		t.Fatalf("ERR node config didn't match")
	}

	rawConfig, err = generateConfig(0, 1)
	testRawConfig = `---
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
...`
	if strings.Compare(testRawConfig, string(rawConfig)) != 0 {
		t.Fatalf("ERR node config didn't match")
	}
}
