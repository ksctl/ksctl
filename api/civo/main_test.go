package civo

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	util "github.com/kubesimplify/ksctl/api/utils"
	"github.com/stretchr/testify/assert"
)

func TestFetchAPIKey(T *testing.T) {
	apikey := fetchAPIKey()

	if fmt.Sprintf("%T", apikey) != "string" || len(apikey) != 0 {
		T.Fatalf("Invalid Return type or APIKey already present")
	}
}

func TestIsValidNodeSize(T *testing.T) {
	validSizes := []string{"g4s.kube.xsmall", "g4s.kube.small", "g4s.kube.medium", "g4s.kube.large", "g4p.kube.small", "g4p.kube.medium", "g4p.kube.large", "g4p.kube.xlarge", "g4c.kube.small", "g4c.kube.medium", "g4c.kube.large", "g4c.kube.xlarge", "g4m.kube.small", "g4m.kube.medium", "g4m.kube.large", "g4m.kube.xlarge"}
	testData := validSizes[rand.Int()%len(validSizes)]
	assert.Equalf(T, true, isValidSizeManaged(testData), "Returns False for valid size")

	assert.Equalf(T, false, isValidSizeManaged("abcd"), "Returns True for invalid node size")
	assert.Equalf(T, false, isValidSizeManaged("kube.small"), "Returns True for invalid node size")
	assert.Equalf(T, false, isValidSizeManaged("g4s.k3s.small"), "Returns True for invalid node size")
}

//TODO: Test ClusterInfoInjecter()

//TODO: Test kubeconfigDeleter()

//Testing of deleteClusterWithID() and DeleteCluster() and CreateCluster() [TODO Need to be done]

func setup() {
	err := os.MkdirAll(util.GetPathCIVO(1, "civo"), 0750)
	if err != nil {
		return
	}
}

func clean() {
	//_ = DeleteCluster(clusterName)
	err := os.RemoveAll(util.GetPathCIVO(1, "civo"))
	if err != nil {
		return
	}
}

func TestIsPresent(t *testing.T) {
	setup()
	present := isPresent("civo", "demo", "LON1")
	assert.Equal(t, false, present, "with no clusters returns true! (false +ve)")
	err := os.Mkdir(util.GetPathCIVO(1, "civo", "demo LON1"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Create(util.GetPathCIVO(1, "civo", "demo LON1", "info.json"))
	if err != nil {
		t.Fatal(err)
	}
	present = isPresent("civo", "demo", "LON1")
	clean()
	assert.Equal(t, true, present, "Failed to detect the cluster (false -ve)")
}
