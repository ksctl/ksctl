package cmd

import "github.com/spf13/cobra"

func verboseFlags() {
	//createClusterAws.Flags().BoolP("verbose", "v", true, "for verbose output")
	createClusterAzure.Flags().BoolP("verbose", "v", true, "for verbose output")
	createClusterCivo.Flags().BoolP("verbose", "v", true, "for verbose output")
	createClusterLocal.Flags().BoolP("verbose", "v", true, "Verbose output")
	createClusterHACivo.Flags().BoolP("verbose", "v", true, "for verbose output")
	createClusterHAAzure.Flags().BoolP("verbose", "v", true, "for verbose output")

	deleteClusterAzure.Flags().BoolP("verbose", "v", true, "for verbose output")
	deleteClusterCivo.Flags().BoolP("verbose", "v", true, "for verbose output")
	deleteClusterHAAzure.Flags().BoolP("verbose", "v", true, "for verbose output")
	deleteClusterHACivo.Flags().BoolP("verbose", "v", true, "for verbose output")
	deleteClusterLocal.Flags().BoolP("verbose", "v", true, "for verbose output")

	addMoreWorkerNodesHAAzure.Flags().BoolP("verbose", "v", true, "for verbose output")
	addMoreWorkerNodesHACivo.Flags().BoolP("verbose", "v", true, "for verbose output")

	deleteNodesHAAzure.Flags().BoolP("verbose", "v", true, "for verbose output")
	deleteNodesHACivo.Flags().BoolP("verbose", "v", true, "for verbose output")

	getClusterCmd.Flags().BoolP("verbose", "v", true, "for verbose output")
	switchCluster.Flags().BoolP("verbose", "v", true, "for verbose output")

	createClusterAzure.Flags().BoolP("approve", "", true, "approval to avoid showMsg")
	createClusterCivo.Flags().BoolP("approve", "", true, "approval to avoid showMsg")
	createClusterLocal.Flags().BoolP("approve", "", true, "approval to avoid showMsg")
	createClusterHACivo.Flags().BoolP("approve", "", true, "approval to avoid showMsg")
	createClusterHAAzure.Flags().BoolP("approve", "", true, "approval to avoid showMsg")
	deleteClusterAzure.Flags().BoolP("approve", "", true, "approval to avoid showMsg")
	deleteClusterCivo.Flags().BoolP("approve", "", true, "approval to avoid showMsg")
	deleteClusterHAAzure.Flags().BoolP("approve", "", true, "approval to avoid showMsg")
	deleteClusterHACivo.Flags().BoolP("approve", "", true, "approval to avoid showMsg")
	deleteClusterLocal.Flags().BoolP("approve", "", true, "approval to avoid showMsg")
	addMoreWorkerNodesHAAzure.Flags().BoolP("approve", "", true, "approval to avoid showMsg")
	addMoreWorkerNodesHACivo.Flags().BoolP("approve", "", true, "approval to avoid showMsg")
	deleteNodesHAAzure.Flags().BoolP("approve", "", true, "approval to avoid showMsg")
	deleteNodesHACivo.Flags().BoolP("approve", "", true, "approval to avoid showMsg")
}

func clusterNameFlag(f *cobra.Command) {
	f.Flags().StringVarP(&clusterName, "name", "n", "demo", "Cluster Name") // keep it same for all
}

func nodeSizeManagedFlag(f *cobra.Command) {
	f.Flags().StringVarP(&nodeSizeMP, "nodeSizeMP", "", "", "Node size of managed cluster nodes")
}

func nodeSizeCPFlag(f *cobra.Command) {
	f.Flags().StringVarP(&nodeSizeCP, "nodeSizeCP", "", "", "Node size of self-managed controlplane nodes")
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
