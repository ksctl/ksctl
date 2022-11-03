package local

import (
	"github.com/stretchr/testify/assert"
	"os"
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

func cleanup() {
	//_ = DeleteCluster(clusterName)
	err := os.RemoveAll(GetPath())
	if err != nil {
		return
	}
}

func setup() {
	err := os.MkdirAll(GetPath(), 0750)
	if err != nil {
		return
	}
}

func TestIsPresent(t *testing.T) {
	setup()
	present := isPresent("demo")
	assert.Equal(t, false, present, "with no clusters returns true! (false +ve)")
	//abcd := ClusterInfoInjecter("demo", 1)
	//if err := CreateCluster(abcd); err != nil {
	//	fmt.Println("[DEBUG] failed to create cluster: ", err.Error())
	//	t.Fatalf("Unable to create cluster ENV not supported!")
	//}

	// create the folder
	err := os.Mkdir(GetPath("demo"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Create(GetPath("demo", "info"))
	if err != nil {
		t.Fatal(err)
	}

	present = isPresent("demo")
	cleanup()
	assert.Equal(t, true, present, "Failed to detect the cluster (false -ve)")
}
