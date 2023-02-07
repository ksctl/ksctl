/*
Kubesimplify
Credit to @civo
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

package civo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	util "github.com/kubesimplify/ksctl/api/utils"
)

type CivoProvider struct {
	ClusterName string
	APIKey      string
	HACluster   bool
	Region      string
	Spec        util.Machine
	Application string
	CNIPlugin   string
}

func Credentials() bool {

	apikey := ""
	fmt.Println("Enter your API-TOKEN-KEY: ")
	_, err := fmt.Scan(&apikey)
	if err != nil {
		panic(err.Error())
	}

	apiStore := util.CivoCredential{
		Token: apikey,
	}

	err = saveCred(apiStore)

	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

func getCred() (configStore util.CivoCredential, err error) {

	fileBytes, err := os.ReadFile(util.GetPath(util.CREDENTIAL_PATH, "civo"))

	if err != nil {
		return
	}

	err = json.Unmarshal(fileBytes, &configStore)

	if err != nil {
		return
	}

	return
}

func saveCred(configStore util.CivoCredential) error {

	storeBytes, err := json.Marshal(configStore)
	if err != nil {
		return err
	}
	_, err = os.Create(util.GetPath(util.CREDENTIAL_PATH, "civo"))
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = ioutil.WriteFile(util.GetPath(util.CREDENTIAL_PATH, "civo"), storeBytes, 0640)
	if err != nil {
		return err
	}
	log.Println("💾 configuration")
	return nil
}

// fetchAPIKey returns the API key from the cred/civo file store
func fetchAPIKey() string {

	token, err := getCred()

	if err != nil {
		return ""
	}
	return token.Token
}

// isPresent Checks whether the cluster to create is already had been created
func isPresent(offering, clusterName, Region string) bool {
	_, err := os.ReadFile(util.GetPath(util.CLUSTER_PATH, "civo", offering, clusterName+" "+Region, "info.json"))
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// cleanup called when error is encountered during creation og cluster
func cleanup(provider CivoProvider) error {
	log.Println("[ERR] Cannot continue 😢")
	return haDeleteClusterHandler(provider.ClusterName, provider.Region, false)
}

func validationOfArguments(name, region string) error {

	if !util.IsValidRegionCIVO(region) {
		return fmt.Errorf("🚩 REGION")
	}

	if !util.IsValidName(name) {
		return fmt.Errorf("🚩 NAME FORMAT")
	}

	return nil
}

func (provider CivoProvider) CreateCluster() error {
	if provider.HACluster {
		if err := haCreateClusterHandler(provider.ClusterName, provider.Region, provider.Spec.Disk,
			provider.Spec.HAControlPlaneNodes, provider.Spec.HAWorkerNodes); err != nil {
			_ = cleanup(provider)
			return err
		}
		return nil
	}
	payload := ClusterInfoInjecter(provider.ClusterName, provider.Region, provider.Spec.Disk, provider.Spec.ManagedNodes, provider.Application, provider.CNIPlugin)
	return managedCreateClusterHandler(payload)
}

func (provider CivoProvider) DeleteCluster() error {
	if provider.HACluster {
		return haDeleteClusterHandler(provider.ClusterName, provider.Region, true)
	}
	return managedDeleteClusterHandler(provider.ClusterName, provider.Region)
}

func (provider CivoProvider) SwitchContext() error {
	switch provider.HACluster {
	case true:
		if isPresent("ha", provider.ClusterName, provider.Region) {
			var printKubeconfig util.PrinterKubeconfigPATH
			printKubeconfig = printer{ClusterName: provider.ClusterName, Region: provider.Region}
			printKubeconfig.Printer(true, 0)
			return nil
		}
	case false:
		if isPresent("managed", provider.ClusterName, provider.Region) {
			var printKubeconfig util.PrinterKubeconfigPATH
			printKubeconfig = printer{ClusterName: provider.ClusterName, Region: provider.Region}
			printKubeconfig.Printer(false, 0)
			return nil
		}
	}
	return fmt.Errorf("ERR Cluster not found")
}

// this will be made available to each create functional calls

type printer struct {
	ClusterName string
	Region      string
}

// Printer to print the KUBECONFIG ENV setter command
// isHA: whether the cluster created is HA type or not
// operation: 0 for created cluster operation and 1 for deleted cluster operation
func (p printer) Printer(isHA bool, operation int) {
	// FIXME: add platform dependent code missing windows env set
	switch operation {
	case 0:
		fmt.Printf("\n\033[33;40mTo use this cluster set this environment variable\033[0m\n\n")
		if isHA {
			fmt.Println(fmt.Sprintf("export KUBECONFIG='%s'\n", util.GetPath(util.CLUSTER_PATH, "civo", "ha", p.ClusterName+" "+p.Region, "config")))
		} else {
			fmt.Println(fmt.Sprintf("export KUBECONFIG='%s'\n", util.GetPath(util.CLUSTER_PATH, "civo", "managed", p.ClusterName+" "+p.Region, "config")))
		}
	case 1:
		fmt.Printf("\n\033[33;40mUse the following command to unset KUBECONFIG\033[0m\n\n")
		fmt.Println(fmt.Sprintf("unset KUBECONFIG\n"))
	}
	fmt.Println()
}
