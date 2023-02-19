package azure

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

import (
	"context"
	"log"
)

func haCreateClusterHandler(ctx context.Context, obj *AzureProvider) error {
	log.Println("start creating virtual machine...")
	resourceGroup, err := obj.CreateResourceGroup(ctx)
	if err != nil {
		log.Fatalf("cannot create resource group:%+v", err)
	}
	log.Printf("Created resource group: %s", *resourceGroup.ID)

	virtualNetwork, err := obj.CreateVirtualNetwork(ctx)
	if err != nil {
		log.Fatalf("cannot create virtual network:%+v", err)
	}
	log.Printf("Created virtual network: %s", *virtualNetwork.ID)

	subnet, err := obj.CreateSubnet(ctx)
	if err != nil {
		log.Fatalf("cannot create subnet:%+v", err)
	}
	log.Printf("Created subnet: %s", *subnet.ID)

	publicIP, err := obj.CreatePublicIP(ctx, obj.ClusterName+"-pub-ip")
	if err != nil {
		log.Fatalf("cannot create public IP address:%+v", err)
	}
	log.Printf("Created public IP address: %s", *publicIP.ID)

	// network security group
	nsg, err := obj.CreateNSG(ctx, obj.ClusterName+"-nsg")
	if err != nil {
		log.Fatalf("cannot create network security group:%+v", err)
	}
	log.Printf("Created network security group: %s", *nsg.ID)

	networkInterface, err := obj.CreateNetworkInterface(ctx, obj.Config.ResourceGroupName, obj.ClusterName+"-nic", *subnet.ID, *publicIP.ID, *nsg.ID)
	if err != nil {
		log.Fatalf("cannot create network interface:%+v", err)
	}
	log.Printf("Created network interface: %s", *networkInterface.ID)

	networkInterfaceID := networkInterface.ID
	virtualMachine, err := obj.CreateVM(ctx, obj.ClusterName+"-cp-1", *networkInterfaceID, obj.ClusterName+"-disk")
	if err != nil {
		log.Fatalf("cannot create virual machine:%+v", err)
	}
	log.Printf("Created network virual machine: %s", *virtualMachine.ID)

	log.Println("Virtual machine created successfully")
	return nil
}

func haDeleteClusterHandler(ctx context.Context, obj *AzureProvider) error {
	log.Println("start deleting virtual machine...")
	err := obj.DeleteVM(ctx, obj.ClusterName+"-cp-1")
	if err != nil {
		log.Fatalf("cannot delete virtual machine:%+v", err)
	}
	log.Println("deleted virtual machine")

	err = obj.DeleteDisk(ctx, obj.ClusterName+"-disk")
	if err != nil {
		log.Fatalf("cannot delete disk:%+v", err)
	}
	log.Println("deleted disk")

	err = obj.DeleteNetworkInterface(ctx, obj.ClusterName+"-nic")
	if err != nil {
		log.Fatalf("cannot delete network interface:%+v", err)
	}
	log.Println("deleted network interface")

	err = obj.DeleteNSG(ctx, obj.ClusterName+"-nsg")
	if err != nil {
		log.Fatalf("cannot delete network security group:%+v", err)
	}
	log.Println("deleted network security group")

	err = obj.DeletePublicIP(ctx, obj.ClusterName+"-pub-ip")
	if err != nil {
		log.Fatalf("cannot delete public IP address:%+v", err)
	}
	log.Println("deleted public IP address")

	err = obj.DeleteSubnet(ctx)
	if err != nil {
		log.Fatalf("cannot delete subnet:%+v", err)
	}
	log.Println("deleted subnet")

	err = obj.DeleteVirtualNetwork(ctx)
	if err != nil {
		log.Fatalf("cannot delete virtual network:%+v", err)
	}
	log.Println("deleted virtual network")

	err = obj.DeleteResourceGroup(ctx)
	if err != nil {
		log.Fatalf("cannot delete resource group:%+v", err)
	}
	log.Println("deleted resource group")
	log.Println("success deleted virtual machine.")
	return nil
}
