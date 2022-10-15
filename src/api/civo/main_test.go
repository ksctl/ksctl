package civo

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

//TODO: By Default fetchAPIKey() should return ""
func TestFetchAPIKey(T *testing.T) {
	apikey := fetchAPIKey()

	if fmt.Sprintf("%T", apikey) != "string" || len(apikey) != 0 {
		T.Fatalf("Invalid Return type or APIKey already present")
	}
}

//TODO: Test isValidRegion()
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
		if expRet != isValidRegion(reg) {
			T.Fatalf("Invalid Region Code is passed!")
		}
	}
}

//TODO: Test getUserName()
func TestGetUserName(T *testing.T) {
	usrCmd := exec.Command("whoami")

	output, err := usrCmd.Output()
	if err != nil {
		T.Fatalf("Command exec failed")
	}
	userName := strings.Trim(string(output), "\n")
	if strings.Compare(userName, getUserName()) != 0 {
		T.Fatalf("Couldn't retrieve the corrent username")
	}
}

//TODO: Test ClusterInfoInjecter()

//TODO: Test kubeconfigDeleter()

//Testing of deleteClusterWithID() and DeleteCluster() and CreateCluster() [TODO Need to be done]
