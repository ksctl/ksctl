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
	"os"
	"runtime"

	log "github.com/kubesimplify/ksctl/api/logger"

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

// Credentials accept the api_token for CIVO auth and authorization from user
func Credentials(logger log.Logger) bool {

	logger.Print("Enter your API-TOKEN-KEY 👇")
	apikey, err := util.UserInputCredentials(logger)
	if err != nil {
		panic(err.Error())
	}

	apiStore := util.CivoCredential{
		Token: apikey,
	}

	err = util.SaveCred(logger, apiStore, "civo")

	if err != nil {
		logger.Err(err.Error())
		return false
	}
	return true
}

// fetchAPIKey returns the api_token from the cred/civo.json file store
func fetchAPIKey(logger log.Logger) string {

	civoToken := os.Getenv("CIVO_TOKEN")
	if civoToken != "" {
		return civoToken
	}
	logger.Warn("environment vars not set: CIVO_TOKEN")

	token, err := util.GetCred(logger, "civo")
	if err != nil {
		return ""
	}
	return token["token"]
}

// isPresent Checks whether the cluster to create is already present
func isPresent(offering, clusterName, Region string) bool {
	_, err := os.ReadFile(util.GetPath(util.CLUSTER_PATH, "civo", offering, clusterName+" "+Region, "info.json"))
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// cleanup called when error is encountered during creation during cluster creation
func cleanup(logging log.Logger, provider CivoProvider) error {
	logging.Err("Cannot continue 😢")
	return haDeleteClusterHandler(logging, provider.ClusterName, provider.Region, false)
}

// validationOfArguments is name and region specified valid
func validationOfArguments(name, region string) error {

	if !util.IsValidRegionCIVO(region) {
		return fmt.Errorf("REGION")
	}

	if !util.IsValidName(name) {
		return fmt.Errorf("NAME FORMAT")
	}

	return nil
}

// CreateCluster calls the helper functions for cluster creation
// based on the flag `HACluster` whether to delete managed cluster or HA type cluster
// FIXME: Ingress or Loadbalancer is not working as expected!!
func (provider CivoProvider) CreateCluster(logging log.Logger) error {
	if provider.HACluster {
		if err := haCreateClusterHandler(logging, provider.ClusterName, provider.Region, provider.Spec.Disk,
			provider.Spec.HAControlPlaneNodes, provider.Spec.HAWorkerNodes); err != nil {
			_ = cleanup(logging, provider)
			return err
		}
		return nil
	}
	payload := ClusterInfoInjecter(logging, provider.ClusterName, provider.Region, provider.Spec.Disk, provider.Spec.ManagedNodes, provider.Application, provider.CNIPlugin)
	err := managedCreateClusterHandler(logging, payload) // FIXME: no cleanup defined
	if err != nil {
		logging.Err("CLEANUP TRIGGERED!: failed to create")
		_ = managedDeleteClusterHandler(logging, provider.ClusterName, provider.Region)
		return err
	}
	return err
}

// DeleteCluster calls the helper functions for cluster deletion
// based on the flag `HACluster` whether to delete managed cluster or HA type cluster
func (provider CivoProvider) DeleteCluster(logging log.Logger) error {
	if provider.HACluster {
		return haDeleteClusterHandler(logging, provider.ClusterName, provider.Region, true)
	}
	return managedDeleteClusterHandler(logging, provider.ClusterName, provider.Region)
}

// SwitchContext provides the export command for switching to specific provider's cluster
func (provider CivoProvider) SwitchContext(logging log.Logger) error {
	switch provider.HACluster {
	case true:
		if isPresent("ha", provider.ClusterName, provider.Region) {
			var printKubeconfig util.PrinterKubeconfigPATH
			printKubeconfig = printer{ClusterName: provider.ClusterName, Region: provider.Region}
			printKubeconfig.Printer(logging, true, 0)
			return nil
		}
	case false:
		if isPresent("managed", provider.ClusterName, provider.Region) {
			var printKubeconfig util.PrinterKubeconfigPATH
			printKubeconfig = printer{ClusterName: provider.ClusterName, Region: provider.Region}
			printKubeconfig.Printer(logging, false, 0)
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
func (p printer) Printer(logging log.Logger, isHA bool, operation int) {
	preFix := "export "
	if runtime.GOOS == "windows" {
		preFix = "$Env:"
	}
	switch operation {
	case 0:
		logging.Note("To use this cluster set this environment variable")
		if isHA {
			logging.Print(fmt.Sprintf("%sKUBECONFIG=\"%s\"\n", preFix, util.GetPath(util.CLUSTER_PATH, "civo", "ha", p.ClusterName+" "+p.Region, "config")))
		} else {
			logging.Print(fmt.Sprintf("%sKUBECONFIG=\"%s\"\n", preFix, util.GetPath(util.CLUSTER_PATH, "civo", "managed", p.ClusterName+" "+p.Region, "config")))
		}
	case 1:
		logging.Note("Use the following command to unset KUBECONFIG")
		if runtime.GOOS == "windows" {
			logging.Print(fmt.Sprintf("%sKUBECONFIG=\"\"\n", preFix))
		} else {
			logging.Print("unset KUBECONFIG")
		}
	}
}
