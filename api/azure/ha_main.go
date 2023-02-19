package azure

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

import (
	"context"
	"fmt"
	"log"
)

func haCreateClusterHandler(ctx context.Context, obj *AzureProvider) error {
	log.Println("Started to Create your HA cluster on Azure provider...")
	defer obj.ConfigWriter("ha")

	_, err := obj.CreateResourceGroup(ctx)
	if err != nil {
		return err
	}

	err = obj.UploadSSHKey(ctx)
	if err != nil {
		return err
	}

	err = obj.createLoadBalancer(ctx)
	if err != nil {
		return err
	}

	log.Println("Your cluster is now ready")
	return nil
}

func haDeleteClusterHandler(ctx context.Context, obj *AzureProvider) error {
	log.Println("start deleting the cluster...")

	obj.Config.ResourceGroupName = obj.ClusterName + "-ksctl"
	err := obj.ConfigReader("ha")
	if err != nil {
		return fmt.Errorf("Unable to read configuration: %v", err)
	}

	err = obj.DeleteAllVMs(ctx)
	if err != nil {
		return err
	}

	err = obj.DeleteAllDisks(ctx)
	if err != nil {
		return err
	}

	err = obj.DeleteAllNetworkInterface(ctx)
	if err != nil {
		return err
	}

	err = obj.DeleteAllNSG(ctx)
	if err != nil {
		return err
	}

	err = obj.DeleteAllPublicIP(ctx)
	if err != nil {
		return err
	}

	// obj.Config.VirtualNetworkName = obj.ClusterName + "-vn" // TODO: remove this
	err = obj.DeleteSubnet(ctx, obj.Config.SubnetName)
	if err != nil {
		return err
	}

	err = obj.DeleteVirtualNetwork(ctx)
	if err != nil {
		return err
	}

	err = obj.DeleteSSHKeyPair(ctx)
	if err != nil {
		return err
	}

	err = obj.DeleteResourceGroup(ctx)
	if err != nil {
		return err
	}

	return nil
}
