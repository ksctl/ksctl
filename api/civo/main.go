/*
Kubesimplify
Credit to @civo
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

package civo

import (
	"fmt"
	"log"
	"os"
	"strings"

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
	file, err := os.OpenFile(util.GetPathCIVO(0), os.O_WRONLY, 0640)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	apikey := ""
	fmt.Println("Enter your API-TOKEN-KEY: ")
	_, err = fmt.Scan(&apikey)
	if err != nil {
		panic(err.Error())
	}

	_, err = file.Write([]byte(fmt.Sprintf("API-TOKEN-Key: %s", apikey)))
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// fetchAPIKey returns the API key from the cred/civo file store
func fetchAPIKey() string {

	file, err := os.ReadFile(util.GetPathCIVO(0))
	if err != nil {
		return ""
	}
	if len(file) == 0 {
		return ""
	}

	return strings.Split(string(file), " ")[1]
}

// isPresent Checks whether the cluster to create is already had been created
func isPresent(offering, clusterName, Region string) bool {
	// FIXME: the ha and managed have 2 different type of config storeage
	_, err := os.ReadFile(util.GetPathCIVO(1, offering, clusterName+" "+Region, "info.json"))
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// cleanup called when error is encountered during creation og cluster
func cleanup(provider CivoProvider) error {
	log.Println("[ERR] Cannot continue 😢")
	return provider.DeleteCluster()
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
		if isPresent("ha-civo", provider.ClusterName, provider.Region) {
			var printKubeconfig util.PrinterKubeconfigPATH
			printKubeconfig = printer{ClusterName: provider.ClusterName, Region: provider.Region}
			printKubeconfig.Printer(true, 0)
			return nil
		}
	case false:
		if isPresent("civo", provider.ClusterName, provider.Region) {
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

func (p printer) Printer(ha bool, a int) {
	switch a {
	case 0:
		fmt.Printf("\n\033[33;40mTo use this cluster set this environment variable\033[0m\n\n")
		if ha {
			fmt.Println(fmt.Sprintf("export KUBECONFIG='%s'\n", util.GetPathCIVO(1, "ha-civo", p.ClusterName+" "+p.Region, "config")))
		} else {
			fmt.Println(fmt.Sprintf("export KUBECONFIG='%s'\n", util.GetPathCIVO(1, "civo", p.ClusterName+" "+p.Region, "config")))
		}
	case 1:
		fmt.Printf("\n\033[33;40mUse the following command to unset KUBECONFIG\033[0m\n\n")
		fmt.Println(fmt.Sprintf("unset KUBECONFIG\n"))
	}
	fmt.Println()
}
