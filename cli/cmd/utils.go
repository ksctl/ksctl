package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kubesimplify/ksctl/api/utils"
	"github.com/spf13/cobra"
)

func createManaged(approval bool) {
	cli.Client.Metadata.ManagedNodeType = nodeSizeMP
	cli.Client.Metadata.NoMP = noMP

	cli.Client.Metadata.ClusterName = clusterName
	cli.Client.Metadata.K8sDistro = distro
	cli.Client.Metadata.K8sVersion = k8sVer
	cli.Client.Metadata.Region = region

	cli.Client.Metadata.CNIPlugin = cni
	cli.Client.Metadata.Applications = apps
	if err := createApproval(approval); err != nil {
		cli.Client.Storage.Logger().Err(err.Error())
		os.Exit(1)
	}

	stat, err := controller.CreateManagedCluster(&cli.Client)
	if err != nil {
		cli.Client.Storage.Logger().Err(err.Error())
		os.Exit(1)
	}
	cli.Client.Storage.Logger().Success(stat)
}

func createHA(approval bool) {
	cli.Client.Metadata.IsHA = true

	cli.Client.Metadata.ClusterName = clusterName
	cli.Client.Metadata.K8sDistro = distro
	cli.Client.Metadata.K8sVersion = k8sVer
	cli.Client.Metadata.Region = region

	cli.Client.Metadata.NoCP = noCP
	cli.Client.Metadata.NoWP = noWP
	cli.Client.Metadata.NoDS = noDS

	cli.Client.Metadata.LoadBalancerNodeType = nodeSizeLB
	cli.Client.Metadata.ControlPlaneNodeType = nodeSizeCP
	cli.Client.Metadata.WorkerPlaneNodeType = nodeSizeWP
	cli.Client.Metadata.DataStoreNodeType = nodeSizeDS

	cli.Client.Metadata.CNIPlugin = cni
	cli.Client.Metadata.Applications = apps

	if err := createApproval(approval); err != nil {
		cli.Client.Storage.Logger().Err(err.Error())
		os.Exit(1)
	}
	stat, err := controller.CreateHACluster(&cli.Client)
	if err != nil {
		cli.Client.Storage.Logger().Err(err.Error())
		os.Exit(1)
	}
	cli.Client.Storage.Logger().Success(stat)
}

func deleteManaged(approval bool) {

	cli.Client.Metadata.ClusterName = clusterName
	cli.Client.Metadata.K8sDistro = distro
	cli.Client.Metadata.Region = region

	if err := deleteApproval(approval); err != nil {
		cli.Client.Storage.Logger().Err(err.Error())
		os.Exit(1)
	}
	stat, err := controller.DeleteManagedCluster(&cli.Client)
	if err != nil {
		cli.Client.Storage.Logger().Err(err.Error())
		os.Exit(1)
	}
	cli.Client.Storage.Logger().Success(stat)
}

func deleteHA(approval bool) {

	cli.Client.Metadata.IsHA = true

	cli.Client.Metadata.ClusterName = clusterName
	cli.Client.Metadata.K8sDistro = distro
	cli.Client.Metadata.Region = region

	if err := deleteApproval(approval); err != nil {
		cli.Client.Storage.Logger().Err(err.Error())
		os.Exit(1)
	}
	stat, err := controller.DeleteHACluster(&cli.Client)
	if err != nil {
		cli.Client.Storage.Logger().Err(err.Error())
		os.Exit(1)
	}
	cli.Client.Storage.Logger().Success(stat)
}

func getRequestPayload() ([]byte, error) {
	a, err := json.MarshalIndent(cli.Client.Metadata, "", " ")
	if err != nil {
		return nil, err
	}
	return a, nil
}

func deleteApproval(showMsg bool) error {

	a, err := getRequestPayload()
	if err != nil {
		return err
	}
	fmt.Println(string(a))

	if !showMsg {
		fmt.Println(fmt.Sprintf("🚨 THIS IS A DESTRUCTIVE STEP MAKE SURE IF YOU WANT TO DELETE THE CLUSTER"))

		fmt.Println("Enter your choice to continue..[y/N]")
		choice := "n"
		unsafe := false
		fmt.Scanf("%s", &choice)
		if strings.Compare("y", choice) == 0 ||
			strings.Compare("yes", choice) == 0 ||
			strings.Compare("Y", choice) == 0 {
			unsafe = true
		}

		if !unsafe {
			return fmt.Errorf("[ksctl] approval cancelled")
		}
	}
	return nil
}

func createApproval(showMsg bool) error {

	a, err := getRequestPayload()
	if err != nil {
		return err
	}
	fmt.Println(string(a))

	if !showMsg {
		fmt.Println(fmt.Sprintf("🚨 THIS IS A CREATION STEP MAKE SURE IF YOU WANT TO CREATE THE CLUSTER"))

		fmt.Println("Enter your choice to continue..[y/N]")
		choice := "n"
		unsafe := false
		fmt.Scanf("%s", &choice)
		if strings.Compare("y", choice) == 0 ||
			strings.Compare("yes", choice) == 0 ||
			strings.Compare("Y", choice) == 0 {
			unsafe = true
		}

		if !unsafe {
			return fmt.Errorf("[ksctl] approval cancelled")
		}
	}
	return nil
}

func clusterNameFlag(f *cobra.Command) {
	f.Flags().StringVarP(&clusterName, "name", "n", "demo", "Cluster Name") // keep it same for all
}

func nodeSizeManagedFlag(f *cobra.Command) {
	f.Flags().StringVarP(&nodeSizeMP, "nodeSizeMP", "", "", "Node size of managed cluster nodes")
}

func nodeSizeCPFlag(f *cobra.Command) {
	f.Flags().StringVarP(&nodeSizeMP, "nodeSizeMP", "", "", "Node size of self-managed controlplane nodes")
}
func nodeSizeWPFlag(f *cobra.Command) {
	f.Flags().StringVarP(&nodeSizeWP, "nodeSizeWP", "", "", "Node size of self-managed workerplane nodes")
}

func nodeSizeDSFlag(f *cobra.Command) {
	f.Flags().StringVarP(&nodeSizeDS, "nodeSizeDS", "", "", "Node size of self-managed datastore nodes")
}

func nodeSizeLBFlag(f *cobra.Command) {
	f.Flags().StringVarP(&nodeSizeLB, "nodeSizeLB", "", "", "Node size of self-managed loadbalancer node")
}

func regionFlag(f *cobra.Command) {
	f.Flags().StringVarP(&region, "region", "r", "", "Region")
}

func appsFlag(f *cobra.Command) {
	f.Flags().StringVarP(&apps, "apps", "", "", "Pre-Installed Applications")
}

func cniFlag(f *cobra.Command) {
	f.Flags().StringVarP(&cni, "cni", "", "", "CNI")
}

func distroFlag(f *cobra.Command) {
	f.Flags().StringVarP(&distro, "distribution", "", "", "Kubernetes Distribution")
}

func k8sVerFlag(f *cobra.Command) {
	f.Flags().StringVarP(&k8sVer, "version", "", "", "Kubernetes Version")
}

func noOfWPFlag(f *cobra.Command) {
	f.Flags().IntVarP(&noWP, "noWP", "", -1, "Number of WorkerPlane Nodes")
}
func noOfCPFlag(f *cobra.Command) {
	f.Flags().IntVarP(&noCP, "noCP", "", -1, "Number of ControlPlane Nodes")
}
func noOfMPFlag(f *cobra.Command) {
	f.Flags().IntVarP(&noMP, "noMP", "", -1, "Number of Managed Nodes")
}
func noOfDSFlag(f *cobra.Command) {
	f.Flags().IntVarP(&noDS, "noDS", "", -1, "Number of DataStore Nodes")
}

func SetDefaults(provider, clusterType string) {
	switch provider + clusterType {
	case utils.CLOUD_LOCAL + utils.CLUSTER_TYPE_MANG:
		if noMP == -1 {
			noMP = 2
		}
		if len(k8sVer) == 0 {
			k8sVer = "1.27.1"
		}
		if len(distro) == 0 {
			distro = utils.K8S_K3S
		}

	case utils.CLOUD_AZURE + utils.CLUSTER_TYPE_MANG:
		if len(nodeSizeMP) == 0 {
			nodeSizeMP = "Standard_DS2_v2"
		}
		if noMP == -1 {
			noMP = 1
		}
		if len(region) == 0 {
			region = "eastus"
		}
		if len(k8sVer) == 0 {
			k8sVer = "1.27"
		}
		if len(distro) == 0 {
			distro = utils.K8S_K3S
		}

	case utils.CLOUD_CIVO + utils.CLUSTER_TYPE_MANG:
		if len(nodeSizeMP) == 0 {
			nodeSizeMP = "g4s.kube.small"
		}
		if noMP == -1 {
			noMP = 1
		}
		if len(region) == 0 {
			region = "LON1"
		}
		if len(k8sVer) == 0 {
			k8sVer = "1.27.1"
		}
		if len(distro) == 0 {
			distro = utils.K8S_K3S
		}

	case utils.CLOUD_AZURE + utils.CLUSTER_TYPE_HA:
		if len(nodeSizeCP) == 0 {
			nodeSizeCP = "Standard_F2s"
		}
		if len(nodeSizeWP) == 0 {
			nodeSizeWP = "Standard_F2s"
		}
		if len(nodeSizeDS) == 0 {
			nodeSizeDS = "Standard_F2s"
		}
		if len(nodeSizeLB) == 0 {
			nodeSizeLB = "Standard_F2s"
		}
		if len(region) == 0 {
			region = "eastus"
		}
		if noWP == -1 {
			noWP = 1
		}
		if noCP == -1 {
			noCP = 3
		}
		if noDS == -1 {
			noDS = 1
		}
		if len(k8sVer) == 0 {
			k8sVer = "1.27.1"
		}
		if len(distro) == 0 {
			distro = utils.K8S_K3S
		}

	case utils.CLOUD_CIVO + utils.CLUSTER_TYPE_HA:
		if len(nodeSizeCP) == 0 {
			nodeSizeCP = "g3.small"
		}
		if len(nodeSizeWP) == 0 {
			nodeSizeWP = "g3.large"
		}
		if len(nodeSizeDS) == 0 {
			nodeSizeDS = "g3.small"
		}
		if len(nodeSizeLB) == 0 {
			nodeSizeLB = "g3.small"
		}
		if len(region) == 0 {
			region = "LON1s"
		}
		if noWP == -1 {
			noWP = 1
		}
		if noCP == -1 {
			noCP = 3
		}
		if noDS == -1 {
			noDS = 1
		}
		if len(k8sVer) == 0 {
			k8sVer = "1.27.1"
		}
		if len(distro) == 0 {
			distro = utils.K8S_K3S
		}
	}
}
