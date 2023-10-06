package cmd

import "github.com/spf13/cobra"

func verboseFlags() {
	msgVerbose := "for verbose output"
	msgApproval := "approval to avoid showMsg"

	//createClusterAws.Flags().BoolP("verbose", "v", true, "for verbose output")
	createClusterAzure.Flags().BoolP("verbose", "v", true, msgVerbose)
	createClusterCivo.Flags().BoolP("verbose", "v", true, msgVerbose)
	createClusterLocal.Flags().BoolP("verbose", "v", true, msgVerbose)
	createClusterHACivo.Flags().BoolP("verbose", "v", true, msgVerbose)
	createClusterHAAzure.Flags().BoolP("verbose", "v", true, msgVerbose)

	deleteClusterAzure.Flags().BoolP("verbose", "v", true, msgVerbose)
	deleteClusterCivo.Flags().BoolP("verbose", "v", true, msgVerbose)
	deleteClusterHAAzure.Flags().BoolP("verbose", "v", true, msgVerbose)
	deleteClusterHACivo.Flags().BoolP("verbose", "v", true, msgVerbose)
	deleteClusterLocal.Flags().BoolP("verbose", "v", true, msgVerbose)

	addMoreWorkerNodesHAAzure.Flags().BoolP("verbose", "v", true, msgVerbose)
	addMoreWorkerNodesHACivo.Flags().BoolP("verbose", "v", true, msgVerbose)

	deleteNodesHAAzure.Flags().BoolP("verbose", "v", true, msgVerbose)
	deleteNodesHACivo.Flags().BoolP("verbose", "v", true, msgVerbose)

	getClusterCmd.Flags().BoolP("verbose", "v", true, msgVerbose)
	switchCluster.Flags().BoolP("verbose", "v", true, msgVerbose)

	createClusterAzure.Flags().BoolP("approve", "", true, msgApproval)
	createClusterCivo.Flags().BoolP("approve", "", true, msgApproval)
	createClusterLocal.Flags().BoolP("approve", "", true, msgApproval)
	createClusterHACivo.Flags().BoolP("approve", "", true, msgApproval)
	createClusterHAAzure.Flags().BoolP("approve", "", true, msgApproval)
	deleteClusterAzure.Flags().BoolP("approve", "", true, msgApproval)
	deleteClusterCivo.Flags().BoolP("approve", "", true, msgApproval)
	deleteClusterHAAzure.Flags().BoolP("approve", "", true, msgApproval)
	deleteClusterHACivo.Flags().BoolP("approve", "", true, msgApproval)
	deleteClusterLocal.Flags().BoolP("approve", "", true, msgApproval)
	addMoreWorkerNodesHAAzure.Flags().BoolP("approve", "", true, msgApproval)
	addMoreWorkerNodesHACivo.Flags().BoolP("approve", "", true, msgApproval)
	deleteNodesHAAzure.Flags().BoolP("approve", "", true, msgApproval)
	deleteNodesHACivo.Flags().BoolP("approve", "", true, msgApproval)
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
