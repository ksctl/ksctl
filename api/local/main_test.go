package local

import (
	"os"
	"strings"
	"testing"

	util "github.com/kubesimplify/ksctl/api/utils"
	"github.com/stretchr/testify/assert"
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

func cleanup() {
	err := os.RemoveAll(util.GetPath(util.OTHER_PATH, "local"))
	if err != nil {
		return
	}
}

func setup() {
	err := os.MkdirAll(util.GetPath(util.OTHER_PATH, "local"), 0750)
	if err != nil {
		return
	}
}

func TestIsPresent(t *testing.T) {
	setup()
	t.Cleanup(func() {
		cleanup()
	})
	present := isPresent("demo")
	assert.Equal(t, false, present, "with no clusters returns true! (false +ve)")

	// create the folder
	err := os.Mkdir(util.GetPath(util.OTHER_PATH, "local", "demo"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Create(util.GetPath(util.OTHER_PATH, "local", "demo", "info"))
	if err != nil {
		t.Fatal(err)
	}

	present = isPresent("demo")
	assert.Equal(t, true, present, "Failed to detect the cluster (false -ve)")
}
