package ha_civo

import (
	"fmt"
	"github.com/kubesimplify/ksctl/src/api/payload"
	"runtime"
)

//----------------
// ALPHA VERSION
//----------------
import (
	"github.com/civo/civogo"
	"os"
	"strings"
)

// all the configs are present in .ksctl
// want to save the config to ~/.ksctl/config/ha-civo/<Cluster Name> <Region>/*

// TODO: getKubeconfig() fix the path
func getKubeconfig(params ...string) string {
	var ret string

	if runtime.GOOS == "windows" {
		ret = fmt.Sprintf("%s\\.ksctl\\config\\civo", payload.GetUserName())
		for _, item := range params {
			ret += "\\" + item
		}
	} else {
		ret = fmt.Sprintf("%s/.ksctl/config/civo", payload.GetUserName())
		for _, item := range params {
			ret += "/" + item
		}
	}
	return ret
}

func getCredentials() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s\\.ksctl\\cred\\civo", payload.GetUserName())
	} else {
		return fmt.Sprintf("%s/.ksctl/cred/civo", payload.GetUserName())
	}
}

// GetPath use this in every function and differentiate the logic by using if-else
// flag is used to indicate 1 -> KUBECONFIG, 0 -> CREDENTIALS
func GetPath(flag int8, params ...string) string {
	switch flag {
	case 1:
		return getKubeconfig(params...)
	case 0:
		return getCredentials()
	default:
		return ""
	}
}

// fetchAPIKey returns the API key from the cred/civo file store
func fetchAPIKey() string {

	file, err := os.ReadFile(GetPath(0))
	if err != nil {
		return ""
	}
	if len(file) == 0 {
		return ""
	}

	return strings.Split(string(file), " ")[1]
}

func demoScript() string {
	return `#!/bin/bash
echo "Hello"
sudo echo "Demo1" > /home/test.txt
`
}

func CreateVM() {
	var cargo payload.CivoProvider = payload.CivoProvider{Region: "LON1", APIKey: fetchAPIKey()}
	client, err := civogo.NewClient(cargo.APIKey, cargo.Region)
	defaultNetwork, err := client.GetDefaultNetwork()
	if err != nil {
		panic(err.Error())
	}

	abcd := &civogo.InstanceConfig{Hostname: "demo", Region: cargo.Region, Count: 1, NetworkID: defaultNetwork.ID, Script: demoScript()}

	instance, err := client.CreateInstance(abcd)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Password", instance.InitialPassword)
	fmt.Println("Public IP", instance.PublicIP)
	//for true {
	//	getInstance, err := client.GetInstance(instance.ID)
	//	if err != nil {
	//		return
	//	}
	//	fmt.Println(getInstance.Status)
	//
	//
	//}

}
