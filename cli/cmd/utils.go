package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

func createManaged() {
	cli.Client.Metadata.ManagedNodeType = nodeSizeMP
	cli.Client.Metadata.NoMP = noMP

	cli.Client.Metadata.ClusterName = clusterName
	cli.Client.Metadata.K8sDistro = distro
	cli.Client.Metadata.K8sVersion = k8sVer
	cli.Client.Metadata.Region = region

	cli.Client.Metadata.CNIPlugin = cni
	cli.Client.Metadata.Applications = apps

	stat, err := controller.CreateManagedCluster(&cli.Client)
	if err != nil {
		cli.Client.Storage.Logger().Err(err.Error())
		os.Exit(1)
	}
	cli.Client.Storage.Logger().Success(stat)
}

func createHA() {
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

	stat, err := controller.CreateHACluster(&cli.Client)
	if err != nil {
		cli.Client.Storage.Logger().Err(err.Error())
		os.Exit(1)
	}
	cli.Client.Storage.Logger().Success(stat)
}

func deleteManaged() {

	cli.Client.Metadata.ClusterName = clusterName
	cli.Client.Metadata.K8sDistro = distro
	cli.Client.Metadata.Region = region

	stat, err := controller.DeleteManagedCluster(&cli.Client)
	if err != nil {
		cli.Client.Storage.Logger().Err(err.Error())
		os.Exit(1)
	}
	cli.Client.Storage.Logger().Success(stat)
}

func deleteHA() {

	cli.Client.Metadata.IsHA = true

	cli.Client.Metadata.ClusterName = clusterName
	cli.Client.Metadata.K8sDistro = distro
	cli.Client.Metadata.Region = region

	stat, err := controller.DeleteHACluster(&cli.Client)
	if err != nil {
		cli.Client.Storage.Logger().Err(err.Error())
		os.Exit(1)
	}
	cli.Client.Storage.Logger().Success(stat)
}

func clusterNameFlag(f *cobra.Command) {
	f.Flags().StringVarP(&clusterName, "name", "n", "", "Cluster Name")
}

func nodeSizeManagedFlag(f *cobra.Command, defaultVal string) {
	f.Flags().StringVarP(&nodeSizeMP, "nodeSizeMP", "", defaultVal, "Node size of managed cluster nodes")
}

func nodeSizeCPFlag(f *cobra.Command, defaultVal string) {
	f.Flags().StringVarP(&nodeSizeMP, "nodeSizeMP", "", defaultVal, "Node size of self-managed controlplane nodes")
}
func nodeSizeWPFlag(f *cobra.Command, defaultVal string) {
	f.Flags().StringVarP(&nodeSizeWP, "nodeSizeWP", "", defaultVal, "Node size of self-managed workerplane nodes")
}

func nodeSizeDSFlag(f *cobra.Command, defaultVal string) {
	f.Flags().StringVarP(&nodeSizeDS, "nodeSizeDS", "", defaultVal, "Node size of self-managed datastore nodes")
}

func regionFlag(f *cobra.Command, defaultVal string) {
	f.Flags().StringVarP(&region, "region", "r", defaultVal, "Region")
}

func appsFlag(f *cobra.Command, defaultVal string) {
	f.Flags().StringVarP(&apps, "apps", "", defaultVal, "Pre-Installed Applications")
}

func cniFlag(f *cobra.Command, defaultVal string) {
	f.Flags().StringVarP(&cni, "cni", "", defaultVal, "CNI")
}

func distroFlag(f *cobra.Command, defaultVal string) {
	f.Flags().StringVarP(&distro, "distribution", "", defaultVal, "Kubernetes Distribution")
}

func k8sVerFlag(f *cobra.Command, defaultVal string) {
	f.Flags().StringVarP(&k8sVer, "version", "", defaultVal, "Kubernetes Version")
}

func noOfWPFlag(f *cobra.Command, defaultVal int) {
	f.Flags().IntVarP(&noWP, "noWP", "", defaultVal, "Number of WorkerPlane Nodes")
}
func noOfCPFlag(f *cobra.Command, defaultVal int) {
	f.Flags().IntVarP(&noCP, "noCP", "", defaultVal, "Number of ControlPlane Nodes")
}
func noOfMPFlag(f *cobra.Command, defaultVal int) {
	f.Flags().IntVarP(&noMP, "noMP", "", defaultVal, "Number of Managed Nodes")
}
func noOfDSFlag(f *cobra.Command, defaultVal int) {
	f.Flags().IntVarP(&noDS, "noDS", "", defaultVal, "Number of DataStore Nodes")
}
