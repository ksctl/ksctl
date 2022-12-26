package civo

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
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

func TestIsValidRegion(T *testing.T) {
	locationCombinations := map[string]bool{
		"LOC":  false,
		"LON":  false,
		"LON1": true,
		"lon1": false,
		"FRA1": true,
		"NYC":  false,
	}
	for reg, expRet := range locationCombinations {
		if expRet != util.IsValidRegionCIVO(reg) {
			T.Fatalf("Invalid Region Code is passed!")
		}
	}
}

func TestIsValidClusterName(T *testing.T) {
	assert.Equalf(T, true, util.IsValidName("demo"), "Returns True for invalid cluster name")
	assert.Equalf(T, true, util.IsValidName("Dem-o234"), "Returns True for invalid cluster name")
	assert.Equalf(T, true, util.IsValidName("d-234"), "Returns True for invalid cluster name")
	assert.Equalf(T, false, util.IsValidName("234"), "Returns True for invalid cluster name")
	assert.Equalf(T, false, util.IsValidName("-2342"), "Returns True for invalid cluster name")
	assert.Equalf(T, false, util.IsValidName("dscdscsd-#$#$#"), "Returns True for invalid cluster name")
	assert.Equalf(T, false, util.IsValidName("ds@#$#$#"), "Returns True for invalid cluster name")
}

func TestIsValidNodeSize(T *testing.T) {
	validSizes := []string{"g4s.kube.xsmall", "g4s.kube.small", "g4s.kube.medium", "g4s.kube.large", "g4p.kube.small", "g4p.kube.medium", "g4p.kube.large", "g4p.kube.xlarge", "g4c.kube.small", "g4c.kube.medium", "g4c.kube.large", "g4c.kube.xlarge", "g4m.kube.small", "g4m.kube.medium", "g4m.kube.large", "g4m.kube.xlarge"}
	testData := validSizes[rand.Int()%len(validSizes)]
	assert.Equalf(T, true, isValidSize(testData), "Returns False for valid size")

	assert.Equalf(T, false, isValidSize("abcd"), "Returns True for invalid node size")
	assert.Equalf(T, false, isValidSize("kube.small"), "Returns True for invalid node size")
	assert.Equalf(T, false, isValidSize("g4s.k3s.small"), "Returns True for invalid node size")
}

func TestGetUserName(T *testing.T) {
	//usrCmd := exec.Command("whoami")
	//
	//output, err := usrCmd.Output()
	//if err != nil {
	//	T.Fatalf("Command exec failed")
	//}
	//userName := strings.Trim(string(output), "\n")
	if strings.Compare(os.Getenv("HOME"), util.GetUserName()) != 0 {
		T.Fatalf("Couldn't retrieve the corrent username")
	}
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

func cleanup() {
	//_ = DeleteCluster(clusterName)
	err := os.RemoveAll(util.GetPathCIVO(1, "civo"))
	if err != nil {
		return
	}
}

func TestIsPresent(t *testing.T) {
	setup()
	present := isPresent("demo", "LON1")
	assert.Equal(t, false, present, "with no clusters returns true! (false +ve)")
	err := os.Mkdir(util.GetPathCIVO(1, "civo", "demo LON1"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Create(util.GetPathCIVO(1, "civo", "demo LON1", "info"))
	if err != nil {
		t.Fatal(err)
	}
	present = isPresent("demo", "LON1")
	cleanup()
	assert.Equal(t, true, present, "Failed to detect the cluster (false -ve)")
}
