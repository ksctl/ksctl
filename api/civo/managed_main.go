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
	"errors"
	"fmt"

	log "github.com/kubesimplify/ksctl/api/logger"

	"os"
	"strings"
	"time"

	util "github.com/kubesimplify/ksctl/api/utils"

	"github.com/civo/civogo"
)

// configWriterManaged stores the KUBECONFIG
func configWriterManaged(logging log.Logger, kubeconfig, clusterN, region, clusterID string) error {
	// create the necessary folders and files
	clusterFolder := clusterN + " " + region
	err := os.MkdirAll(util.GetPath(util.CLUSTER_PATH, "civo", "managed", clusterFolder), 0750)

	if err != nil && !os.IsExist(err) {
		return err
	}
	_, err = os.Create(util.GetPath(util.CLUSTER_PATH, "civo", "managed", clusterFolder, "config"))
	if err != nil {
		return err
	}
	// write the contents to the req. files
	file, err := os.OpenFile(util.GetPath(util.CLUSTER_PATH, "civo", "managed", clusterFolder, "config"), os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(kubeconfig))
	if err != nil {
		return err
	}

	if err = saveConfigManaged(logging, clusterFolder, ManagedConfig{ClusterID: clusterID, Region: region}); err != nil {
		return err
	}

	return nil
}

type ManagedConfig struct {
	ClusterID string `json:"clusterid"`
	Region    string `json:"region"`
}

// GetConfigManaged fetch the state management file
func GetConfigManaged(clusterName, region string) (configStore ManagedConfig, err error) {

	fileBytes, err := os.ReadFile(util.GetPath(util.CLUSTER_PATH, "civo", "managed", clusterName+" "+region, "info.json"))

	if err != nil {
		return
	}

	err = json.Unmarshal(fileBytes, &configStore)

	if err != nil {
		return
	}

	return
}

// saveConfigManaged update/store the state to state management file
func saveConfigManaged(logging log.Logger, clusterFolder string, configStore ManagedConfig) error {

	storeBytes, err := json.Marshal(configStore)
	if err != nil {
		return err
	}

	err = os.Mkdir(util.GetPath(util.CLUSTER_PATH, "civo", "managed", clusterFolder), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.WriteFile(util.GetPath(util.CLUSTER_PATH, "civo", "managed", clusterFolder, "info.json"), storeBytes, 0640)
	if err != nil {
		return err
	}
	logging.Info("💾 configuration", "")
	return nil
}

// ClusterInfoInjecter Serializes the information which is return as utils.CivoProvider for sending it to API
// {clustername, regionCode, Size of Nodes, No of nodes, Applications(optional), cniPlugion(optional)}
func ClusterInfoInjecter(logging log.Logger, clusterName, reg, size string, noOfNodes int, application, cniPlugin string) CivoProvider {

	if len(application) == 0 {
		application = "Traefik-v2-nodeport,metrics-server" // default: applications
	} else {
		application += ",Traefik-v2-nodeport,metrics-server"
	}
	if len(cniPlugin) == 0 {
		cniPlugin = "flannel" // default: flannel
	}

	spec := CivoProvider{
		ClusterName: clusterName,
		Region:      reg,
		APIKey:      fetchAPIKey(logging),
		HACluster:   false,
		Spec: util.Machine{
			Disk:         size,
			ManagedNodes: noOfNodes,
		},
		Application: application,
		CNIPlugin:   cniPlugin,
	}
	return spec
}

// isValidSize validates the managed K3S civo nodepool nodesize
func isValidSizeManaged(size string) bool {
	validSizes := []string{"g4s.kube.xsmall", "g4s.kube.small", "g4s.kube.medium", "g4s.kube.large", "g4p.kube.small", "g4p.kube.medium", "g4p.kube.large", "g4p.kube.xlarge", "g4c.kube.small", "g4c.kube.medium", "g4c.kube.large", "g4c.kube.xlarge", "g4m.kube.small", "g4m.kube.medium", "g4m.kube.large", "g4m.kube.xlarge"}
	for _, valid := range validSizes {
		if strings.Compare(valid, size) == 0 {
			return true
		}
	}
	fmt.Printf("\n\n\033[34;40mAvailable Node sizes:\n- g4s.kube.xsmall\n- g4s.kube.small\n- g4s.kube.medium\n- g4s.kube.large\n- g4p.kube.small\n- g4p.kube.medium\n- g4p.kube.large\n- g4p.kube.xlarge\n- g4c.kube.small\n- g4c.kube.medium\n- g4c.kube.large\n- g4c.kube.xlarge\n- g4m.kube.small\n- g4m.kube.medium\n- g4m.kube.large\n- g4m.kube.xlarge\033[0m\n")
	return false
}

// CreateCluster creates managed CIVO cluster
func managedCreateClusterHandler(logging log.Logger, civoConfig CivoProvider) error {
	if len(civoConfig.APIKey) == 0 {
		return fmt.Errorf("CREDENTIALS NOT PRESENT")
	}

	if !util.IsValidName(civoConfig.ClusterName) {
		return fmt.Errorf("INVALID CLUSTER NAME")
	}

	if !util.IsValidRegionCIVO(civoConfig.Region) {
		return fmt.Errorf("region code is Invalid")
	}

	if isPresent("managed", civoConfig.ClusterName, civoConfig.Region) {
		return fmt.Errorf("DUPLICATE Cluster")
	}

	if !isValidSizeManaged(civoConfig.Spec.Disk) {
		return fmt.Errorf("INVALID size of node")
	}

	client, err := civogo.NewClient(civoConfig.APIKey, civoConfig.Region)
	if err != nil {
		return err
	}

	defaultNetwork, err := client.GetDefaultNetwork()
	if err != nil {
		return err
	}

	configK8s := &civogo.KubernetesClusterConfig{
		Name:            civoConfig.ClusterName,
		Region:          civoConfig.Region,
		NumTargetNodes:  civoConfig.Spec.ManagedNodes,
		TargetNodesSize: civoConfig.Spec.Disk,
		NetworkID:       defaultNetwork.ID,
		Applications:    civoConfig.Application,
		CNIPlugin:       civoConfig.CNIPlugin,
	}

	resp, err := client.NewKubernetesClusters(configK8s)
	if err != nil {
		if errors.Is(err, civogo.DatabaseKubernetesClusterDuplicateError) {
			return fmt.Errorf("DUPLICATE Cluster FOUND")
		}
		if errors.Is(err, civogo.AuthenticationFailedError) {
			return fmt.Errorf("AUTH FAILED")
		}
		if errors.Is(err, civogo.UnknownError) {
			return fmt.Errorf("UNKNOWN ERR")
		}
	}
	for {
		// clusterDS fetches the current state of kubernetes cluster given its id
		clusterDS, _ := client.GetKubernetesCluster(resp.ID)
		if clusterDS.Ready {
			logging.Info("💻 Booted Instance", civoConfig.ClusterName)
			err := configWriterManaged(logging, clusterDS.KubeConfig, civoConfig.ClusterName, civoConfig.Region, resp.ID)
			if err != nil {
				return err
			}
			logging.Info("✅ Configured", civoConfig.ClusterName)
			var printKubeconfig util.PrinterKubeconfigPATH
			printKubeconfig = printer{ClusterName: civoConfig.ClusterName, Region: civoConfig.Region}
			printKubeconfig.Printer(logging, false, 0)

			break
		}
		logging.Info("🚧 Instance", clusterDS.Status)
		time.Sleep(10 * time.Second)
	}
	logging.Info("Created your managed civo cluster!!🥳 🎉 ")
	return nil
}

// kubeconfigDeleter deletes all configs related to the provided cluster
func kubeconfigDeleter(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	return nil
}

// deleteClusterWithID delete cluster from CIVO by provided regionCode and clusterID
func deleteClusterWithID(logging log.Logger, clusterID, regionCode string) error {
	client, err := civogo.NewClient(fetchAPIKey(logging), regionCode)
	if err != nil {
		return err
	}

	_, err = client.DeleteKubernetesCluster(clusterID)
	if err != nil {
		return err
	}
	logging.Info("Deleting kubernetes cluster", clusterID)
	return nil
}

// DeleteCluster deletes cluster from the given name and region
func managedDeleteClusterHandler(logging log.Logger, name, region string) error {

	// data will contain the saved ClusterID and Region
	data, err := GetConfigManaged(name, region)
	if err != nil {
		return fmt.Errorf("NO matching cluster found")
	}

	if err = deleteClusterWithID(logging, data.ClusterID, data.Region); err != nil {
		return err
	}

	if err := kubeconfigDeleter(util.GetPath(util.CLUSTER_PATH, "civo", "managed", name+" "+region)); err != nil {
		return err
	}

	var printKubeconfig util.PrinterKubeconfigPATH
	printKubeconfig = printer{ClusterName: name, Region: region}
	printKubeconfig.Printer(logging, false, 1)
	return nil
}
