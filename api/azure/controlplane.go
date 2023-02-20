package azure

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

func scriptWithoutCP_1(dbEndpoint, privateIPlb string) string {

	return fmt.Sprintf(`#!/bin/bash
export K3S_DATASTORE_ENDPOINT='%s'
curl -sfL https://get.k3s.io | sh -s - server \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--tls-san %s
`, dbEndpoint, privateIPlb)
}

func scriptWithCP_1() string {
	return `#!/bin/bash
cat /var/lib/rancher/k3s/server/token
`
}

func scriptCP_n(dbEndpoint, privateIPlb, token string) string {
	return fmt.Sprintf(`#!/bin/bash
export SECRET='%s'
export K3S_DATASTORE_ENDPOINT='%s'
curl -sfL https://get.k3s.io | sh -s - server \
	--token=$SECRET \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--tls-san %s
`, token, dbEndpoint, privateIPlb)
}

func scriptKUBECONFIG() string {
	return `#!/bin/bash
cat /etc/rancher/k3s/k3s.yaml`
}

func getControlPlaneFirewallRules() (securityRules []*armnetwork.SecurityRule) {
	securityRules = append(securityRules, &armnetwork.SecurityRule{
		Name: to.Ptr("sample_inbound_6443"),
		Properties: &armnetwork.SecurityRulePropertiesFormat{
			SourceAddressPrefix:      to.Ptr("10.1.0.0/16"),
			SourcePortRange:          to.Ptr("*"),
			DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
			DestinationPortRange:     to.Ptr("6443"),
			Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
			Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
			Priority:                 to.Ptr[int32](100),
			Description:              to.Ptr("sample network security group inbound port 6443"),
			Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
		},
	}, &armnetwork.SecurityRule{
		Name: to.Ptr("sample_inbound_30_to_35k"),
		Properties: &armnetwork.SecurityRulePropertiesFormat{
			SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
			SourcePortRange:          to.Ptr("*"),
			DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
			DestinationPortRange:     to.Ptr("30000-35000"),
			Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
			Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
			Priority:                 to.Ptr[int32](101),
			Description:              to.Ptr("sample network security group inbound port 30000-35000"),
			Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
		},
	})
	return
}

func (obj *AzureProvider) createControlPlane(ctx context.Context, indexOfNode int) error {
	defer obj.ConfigWriter("ha")
	if len(obj.Config.VirtualNetworkName) == 0 || len(obj.Config.SubnetName) == 0 {
		// we need to create the virtual network
		_, err := obj.CreateVirtualNetwork(ctx, obj.ClusterName+"-vnet")
		if err != nil {
			return err
		}

		_, err = obj.CreateSubnet(ctx, obj.ClusterName+"-subnet")
		if err != nil {
			return err
		}
	}

	vmName := fmt.Sprintf("%s-cp-%d", obj.ClusterName, indexOfNode)

	publicIP, err := obj.CreatePublicIP(ctx, vmName+"-pub-ip")
	if err != nil {
		return err
	}
	obj.Config.InfoControlPlanes.PublicIPNames = append(obj.Config.InfoControlPlanes.PublicIPNames, *publicIP.Name)

	// network security group
	if len(obj.Config.InfoControlPlanes.NetworkSecurityGroupName) == 0 {
		nsg, err := obj.CreateNSG(ctx, obj.ClusterName+"-cp-nsg", getControlPlaneFirewallRules())
		if err != nil {
			return err
		}

		obj.Config.InfoControlPlanes.NetworkSecurityGroupName = *nsg.Name
		obj.Config.InfoControlPlanes.NetworkSecurityGroupID = *nsg.ID
	}

	networkInterface, err := obj.CreateNetworkInterface(ctx, obj.Config.ResourceGroupName, vmName+"-nic", obj.Config.SubnetID, *publicIP.ID, obj.Config.InfoControlPlanes.NetworkSecurityGroupID)
	if err != nil {
		return err
	}
	obj.Config.InfoControlPlanes.NetworkInterfaceNames = append(obj.Config.InfoControlPlanes.NetworkInterfaceNames, *networkInterface.Name)

	obj.Config.InfoControlPlanes.Names = append(obj.Config.InfoControlPlanes.Names, vmName)
	obj.Config.InfoControlPlanes.DiskNames = append(obj.Config.InfoControlPlanes.DiskNames, vmName+"-disk")

	_, err = obj.CreateVM(ctx, vmName, *networkInterface.ID, vmName+"-disk", "")
	if err != nil {
		return err
	}
	obj.Config.InfoControlPlanes.PublicIPs = append(obj.Config.InfoControlPlanes.PublicIPs, *publicIP.Properties.IPAddress)
	obj.Config.InfoControlPlanes.PrivateIPs = append(obj.Config.InfoControlPlanes.PrivateIPs, *networkInterface.Properties.IPConfigurations[0].Properties.PrivateIPAddress)
	log.Println("💻 Booted Control plane VM: ", vmName)
	return nil
}
