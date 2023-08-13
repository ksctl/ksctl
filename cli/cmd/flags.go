package cmd

import (
	"github.com/kubesimplify/ksctl/api/utils"
)

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
	nodeSizeManagedFlag(createClusterAzure, "Standard_DS2_v2")
	regionFlag(createClusterAzure, "eastus")
	noOfMPFlag(createClusterAzure, 1)
	k8sVerFlag(createClusterAzure, "1.27")
	distroFlag(createClusterAzure, utils.K8S_K3S)

	// Managed Civo
	clusterNameFlag(createClusterCivo)
	nodeSizeManagedFlag(createClusterCivo, "g4s.kube.small")
	regionFlag(createClusterCivo, "LON1")
	appsFlag(createClusterCivo, "")
	cniFlag(createClusterCivo, "")
	noOfMPFlag(createClusterCivo, 1)
	distroFlag(createClusterCivo, utils.K8S_K3S)
	k8sVerFlag(createClusterCivo, "1.27.1")

	// Managed Local
	clusterNameFlag(createClusterLocal)
	appsFlag(createClusterLocal, "")
	cniFlag(createClusterLocal, "")
	noOfMPFlag(createClusterLocal, 1)
	distroFlag(createClusterLocal, utils.K8S_K3S)
	k8sVerFlag(createClusterLocal, "1.27.1")

	// HA Civo
	clusterNameFlag(createClusterHACivo)
	nodeSizeCPFlag(createClusterHACivo, "g3.small")
	nodeSizeDSFlag(createClusterHACivo, "g3.large")
	nodeSizeWPFlag(createClusterHACivo, "g3.small")
	regionFlag(createClusterHACivo, "LON1")
	appsFlag(createClusterHACivo, "")
	cniFlag(createClusterHACivo, "")
	noOfWPFlag(createClusterHACivo, 1)
	noOfCPFlag(createClusterHACivo, 3)
	noOfDSFlag(createClusterHACivo, 1)
	distroFlag(createClusterHACivo, utils.K8S_K3S)
	k8sVerFlag(createClusterHACivo, "1.27.1")

	// HA Azure
	clusterNameFlag(createClusterHAAzure)
	nodeSizeCPFlag(createClusterHAAzure, "Standard_F2s")
	nodeSizeDSFlag(createClusterHAAzure, "Standard_F2s")
	nodeSizeWPFlag(createClusterHAAzure, "Standard_F2s")
	regionFlag(createClusterHAAzure, "eastus")
	appsFlag(createClusterHAAzure, "")
	cniFlag(createClusterHAAzure, "")
	noOfWPFlag(createClusterHAAzure, 1)
	noOfCPFlag(createClusterHAAzure, 3)
	noOfDSFlag(createClusterHAAzure, 1)
	distroFlag(createClusterHAAzure, utils.K8S_K3S)
	k8sVerFlag(createClusterHAAzure, "1.27.1")

	// Delete commands
	// Managed Local
	clusterNameFlag(deleteClusterLocal)

	// managed Azure
	clusterNameFlag(deleteClusterAzure)
	regionFlag(deleteClusterAzure, "eastus")

	// Managed Civo
	clusterNameFlag(deleteClusterCivo)
	regionFlag(deleteClusterCivo, "LON1")

	// HA Civo
	clusterNameFlag(deleteClusterHACivo)
	regionFlag(deleteClusterHACivo, "LON1")

	// HA Azure
	clusterNameFlag(deleteClusterHAAzure)
	regionFlag(deleteClusterHAAzure, "eastus")
}
