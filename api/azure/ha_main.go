package azure

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

import (
	"context"
	"fmt"
	"os"
	"strings"

	log "github.com/kubesimplify/ksctl/api/logger"
	util "github.com/kubesimplify/ksctl/api/utils"
)

func haCreateClusterHandler(ctx context.Context, logger log.Logger, obj *AzureProvider) error {
	if !util.IsValidName(obj.ClusterName) {
		return fmt.Errorf("invalid cluster name: %v", obj.ClusterName)
	}

	if !isValidNodeSize(obj.Spec.Disk) {
		return fmt.Errorf("node size {%s} is invalid", obj.Spec.Disk)
	}

	if !isValidRegion(obj.Region) {
		return fmt.Errorf("region {%s} is invalid", obj.Region)
	}

	if isPresent("ha", *obj) {
		return fmt.Errorf("cluster already exists: %v", obj.ClusterName)
	}

	logger.Info("Started to Create your HA cluster on Azure provider...", "")
	defer obj.ConfigWriter(logger, "ha")

	_, err := obj.CreateResourceGroup(ctx, logger)
	if err != nil {
		return err
	}

	err = obj.UploadSSHKey(ctx)
	if err != nil {
		return err
	}

	err = obj.createLoadBalancer(logger, ctx)
	if err != nil {
		return err
	}

	err = obj.createDatabase(ctx, logger)
	if err != nil {
		return err
	}

	for i := 0; i < obj.Spec.HAControlPlaneNodes; i++ {
		if err := obj.createControlPlane(ctx, logger, i+1); err != nil {
			return err
		}
	}

	var controlPlaneIPs = make([]string, obj.Spec.HAControlPlaneNodes)
	for i := 0; i < obj.Spec.HAControlPlaneNodes; i++ {
		controlPlaneIPs[i] = obj.Config.InfoControlPlanes.PrivateIPs[i] + ":6443"
	}

	err = obj.ConfigLoadBalancer(logger, controlPlaneIPs)
	if err != nil {
		return err
	}

	token := ""
	mysqlEndpoint := obj.Config.DBEndpoint
	loadBalancerPrivateIP := obj.Config.InfoLoadBalancer.PrivateIP
	for i := 0; i < obj.Spec.HAControlPlaneNodes; i++ {
		if i == 0 {
			err = obj.HelperExecNoOutputControlPlane(logger, obj.Config.InfoControlPlanes.PublicIPs[i], scriptWithoutCP_1(mysqlEndpoint, loadBalancerPrivateIP), true)
			if err != nil {
				return err
			}

			token = obj.GetTokenFromCP_1(logger, obj.Config.InfoControlPlanes.PublicIPs[0])
			if len(token) == 0 {
				return fmt.Errorf("🚨 Cannot retrieve k3s token")
			}
		} else {
			err = obj.HelperExecNoOutputControlPlane(logger, obj.Config.InfoControlPlanes.PublicIPs[i], scriptCP_n(mysqlEndpoint, loadBalancerPrivateIP, token), true)
			if err != nil {
				return err
			}
		}
		logger.Info("✅ Configured", fmt.Sprintf("%s-cp-%d", obj.ClusterName, i+1))
	}

	// Configure the Loadbalancer
	kubeconfig, err := obj.FetchKUBECONFIG(logger, obj.Config.InfoControlPlanes.PublicIPs[0])
	if err != nil {
		return fmt.Errorf("Cannot fetch kubeconfig\n" + err.Error())
	}
	newKubeconfig := strings.Replace(kubeconfig, "127.0.0.1", obj.Config.InfoLoadBalancer.PublicIP, 1)

	newKubeconfig = strings.Replace(newKubeconfig, "default", obj.ClusterName+"-"+obj.Region+"-ha-azure-ksctl", -1)

	err = obj.SaveKubeconfig(logger, newKubeconfig)
	if err != nil {
		return err
	}

	logger.Info("⛓  JOINING WORKER NODES", "")

	for i := 0; i < obj.Spec.HAWorkerNodes; i++ {
		if err := obj.createWorkerPlane(logger, ctx, i+1); err != nil {
			return err
		}
	}
	logger.Info("Created your HA azure cluster!!🥳 🎉 ", "")
	logger.Note("for the very first kubectl API call, do this\n  kubectl cluster-info --insecure-skip-tls-verify\033[0m\nafter this you can proceed with normal operation of the cluster")
	var printKubeconfig util.PrinterKubeconfigPATH
	printKubeconfig = printer{ClusterName: obj.ClusterName, Region: obj.Region, ResourceName: obj.Config.ResourceGroupName}
	printKubeconfig.Printer(true, 0)
	return nil
}

func haDeleteClusterHandler(ctx context.Context, logger log.Logger, obj *AzureProvider, showMsg bool) error {
	if !util.IsValidName(obj.ClusterName) {
		return fmt.Errorf("invalid cluster name: %v", obj.ClusterName)
	}

	if !isValidRegion(obj.Region) {
		return fmt.Errorf("region {%s} is invalid", obj.Region)
	}

	// if !isPresent("ha", *obj) {
	// 	return fmt.Errorf("cluster does not exists: %v", obj.ClusterName)
	// }

	if showMsg {
		logger.Note(fmt.Sprintf(`🚨	THIS IS A DESTRUCTIVE STEP MAKE SURE IF YOU WANT TO DELETE THE CLUSTER '%s'
	`, obj.ClusterName+" "+obj.Config.ResourceGroupName+" "+obj.Region))

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
			return nil
		}
	}

	logger.Info("start deleting the cluster...", "")

	err := obj.ConfigReader(logger, "ha")
	if err != nil {
		return fmt.Errorf("Unable to read configuration: %v", err)
	}

	err = obj.DeleteAllVMs(ctx, logger)
	if err != nil {
		return err
	}

	err = obj.DeleteAllDisks(ctx, logger)
	if err != nil {
		return err
	}

	err = obj.DeleteAllNetworkInterface(ctx, logger)
	if err != nil {
		return err
	}

	err = obj.DeleteAllNSG(ctx, logger)
	if err != nil {
		return err
	}

	err = obj.DeleteAllPublicIP(ctx, logger)
	if err != nil {
		return err
	}

	err = obj.DeleteSubnet(ctx, logger, obj.Config.SubnetName)
	if err != nil {
		return err
	}

	err = obj.DeleteVirtualNetwork(ctx, logger)
	if err != nil {
		return err
	}

	err = obj.DeleteSSHKeyPair(ctx, logger)
	if err != nil {
		return err
	}

	err = obj.DeleteResourceGroup(ctx, logger)
	if err != nil {
		return err
	}
	clusterDir := obj.ClusterName + " " + obj.Config.ResourceGroupName + " " + obj.Region
	if err := os.RemoveAll(util.GetPath(util.CLUSTER_PATH, "azure", "ha", clusterDir)); err != nil {
		return err
	}

	var printKubeconfig util.PrinterKubeconfigPATH
	printKubeconfig = printer{ClusterName: obj.ClusterName, Region: obj.Region, ResourceName: obj.Config.ResourceGroupName}
	printKubeconfig.Printer(false, 1)
	return nil
}
