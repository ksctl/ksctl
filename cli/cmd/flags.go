package cmd

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
}

func argsFlags() {
	// Managed Azure
	clusterNameFlag(createClusterAzure)
	nodeSizeManagedFlag(createClusterAzure)
	regionFlag(createClusterAzure)
	noOfMPFlag(createClusterAzure)
	k8sVerFlag(createClusterAzure)
	distroFlag(createClusterAzure)

	// Managed Civo
	clusterNameFlag(createClusterCivo)
	nodeSizeManagedFlag(createClusterCivo)
	regionFlag(createClusterCivo)
	appsFlag(createClusterCivo)
	cniFlag(createClusterCivo)
	noOfMPFlag(createClusterCivo)
	distroFlag(createClusterCivo)
	k8sVerFlag(createClusterCivo)

	// Managed Local
	clusterNameFlag(createClusterLocal)
	appsFlag(createClusterLocal)
	cniFlag(createClusterLocal)
	noOfMPFlag(createClusterLocal)
	distroFlag(createClusterLocal)
	k8sVerFlag(createClusterLocal)

	// HA Civo
	clusterNameFlag(createClusterHACivo)
	nodeSizeCPFlag(createClusterHACivo)
	nodeSizeDSFlag(createClusterHACivo)
	nodeSizeWPFlag(createClusterHACivo)
	nodeSizeLBFlag(createClusterHACivo)
	regionFlag(createClusterHACivo)
	appsFlag(createClusterHACivo)
	cniFlag(createClusterHACivo)
	noOfWPFlag(createClusterHACivo)
	noOfCPFlag(createClusterHACivo)
	noOfDSFlag(createClusterHACivo)
	distroFlag(createClusterHACivo)
	k8sVerFlag(createClusterHACivo)

	// HA Azure
	clusterNameFlag(createClusterHAAzure)
	nodeSizeCPFlag(createClusterHAAzure)
	nodeSizeDSFlag(createClusterHAAzure)
	nodeSizeWPFlag(createClusterHAAzure)
	nodeSizeLBFlag(createClusterHAAzure)
	regionFlag(createClusterHAAzure)
	appsFlag(createClusterHAAzure)
	cniFlag(createClusterHAAzure)
	noOfWPFlag(createClusterHAAzure)
	noOfCPFlag(createClusterHAAzure)
	noOfDSFlag(createClusterHAAzure)
	distroFlag(createClusterHAAzure)
	k8sVerFlag(createClusterHAAzure)

	// Delete commands
	// Managed Local
	clusterNameFlag(deleteClusterLocal)

	// managed Azure
	clusterNameFlag(deleteClusterAzure)
	regionFlag(deleteClusterAzure)

	// Managed Civo
	clusterNameFlag(deleteClusterCivo)
	regionFlag(deleteClusterCivo)

	// HA Civo
	clusterNameFlag(deleteClusterHACivo)
	regionFlag(deleteClusterHACivo)

	// HA Azure
	clusterNameFlag(deleteClusterHAAzure)
	regionFlag(deleteClusterHAAzure)
}
