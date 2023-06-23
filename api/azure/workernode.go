package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	log "github.com/kubesimplify/ksctl/api/logger"
)

func scriptWP(privateIPlb, token string) string {
	return fmt.Sprintf(`#!/bin/bash
cat <<EOF > worker-setup.sh
#!/bin/bash
curl -sfL https://get.k3s.io | sh -s - agent --token %s --server https://%s:6443
EOF

sudo chmod +x worker-setup.sh
sudo ./worker-setup.sh
`, token, privateIPlb)
}

// TODO: Add more firewall rules
func getWorkerPlaneFirewallRules() (securityRules []*armnetwork.SecurityRule) {
	securityRules = append(securityRules, &armnetwork.SecurityRule{
		Name: to.Ptr("sample_inbound_all_open"),
		Properties: &armnetwork.SecurityRulePropertiesFormat{
			SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
			SourcePortRange:          to.Ptr("*"),
			DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
			DestinationPortRange:     to.Ptr("*"),
			Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
			Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
			Priority:                 to.Ptr[int32](101),
			Description:              to.Ptr("sample network security group inbound port 22"),
			Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
		},
	}, &armnetwork.SecurityRule{
		Name: to.Ptr("sample_inbound_30_to_35k"),
		Properties: &armnetwork.SecurityRulePropertiesFormat{
			SourceAddressPrefix:      to.Ptr("10.0.0.0/8"),
			SourcePortRange:          to.Ptr("*"),
			DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
			DestinationPortRange:     to.Ptr("22"),
			Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
			Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
			Priority:                 to.Ptr[int32](101),
			Description:              to.Ptr("sample network security group inbound port 30000-35000"),
			Direction:                to.Ptr(armnetwork.SecurityRuleDirectionOutbound),
		},
	})
	return
}

// TODO: try to make worker nodes as private instances
func (obj *AzureProvider) createWorkerPlane(logger log.Logger, ctx context.Context, indexOfNode int) error {
	defer obj.ConfigWriter(logger, "ha")
	if len(obj.Config.VirtualNetworkName) == 0 || len(obj.Config.SubnetName) == 0 {
		// we need to create the virtual network
		_, err := obj.CreateVirtualNetwork(ctx, logger, obj.ClusterName+"-vnet")
		if err != nil {
			return err
		}

		_, err = obj.CreateSubnet(ctx, logger, obj.ClusterName+"-subnet")
		if err != nil {
			return err
		}
	}

	vmName := fmt.Sprintf("%s-wp-%d", obj.ClusterName, indexOfNode)

	publicIP, err := obj.CreatePublicIP(ctx, logger, vmName+"-pub-ip")
	if err != nil {
		return err
	}
	obj.Config.InfoWorkerPlanes.PublicIPNames = append(obj.Config.InfoWorkerPlanes.PublicIPNames, *publicIP.Name)

	// network security group
	if len(obj.Config.InfoWorkerPlanes.NetworkSecurityGroupName) == 0 {
		nsg, err := obj.CreateNSG(ctx, logger, obj.ClusterName+"-wp-nsg", getWorkerPlaneFirewallRules())
		if err != nil {
			return err
		}

		obj.Config.InfoWorkerPlanes.NetworkSecurityGroupName = *nsg.Name
		obj.Config.InfoWorkerPlanes.NetworkSecurityGroupID = *nsg.ID
	}

	networkInterface, err := obj.CreateNetworkInterface(ctx, logger, obj.Config.ResourceGroupName, vmName+"-nic", obj.Config.SubnetID, *publicIP.ID, obj.Config.InfoWorkerPlanes.NetworkSecurityGroupID)
	if err != nil {
		return err
	}
	obj.Config.InfoWorkerPlanes.NetworkInterfaceNames = append(obj.Config.InfoWorkerPlanes.NetworkInterfaceNames, *networkInterface.Name)

	obj.Config.InfoWorkerPlanes.Names = append(obj.Config.InfoWorkerPlanes.Names, vmName)
	obj.Config.InfoWorkerPlanes.DiskNames = append(obj.Config.InfoWorkerPlanes.DiskNames, vmName+"-disk")

	_, err = obj.CreateVM(ctx, logger, vmName, *networkInterface.ID, vmName+"-disk", scriptWP(obj.Config.InfoLoadBalancer.PrivateIP, obj.Config.K3sToken))
	if err != nil {
		return err
	}
	obj.Config.InfoWorkerPlanes.PublicIPs = append(obj.Config.InfoWorkerPlanes.PublicIPs, *publicIP.Properties.IPAddress)
	obj.Config.InfoWorkerPlanes.PrivateIPs = append(obj.Config.InfoWorkerPlanes.PrivateIPs, *networkInterface.Properties.IPConfigurations[0].Properties.PrivateIPAddress)
	logger.Info("💻 Booted Worker plane VM: ", vmName)
	return nil
}
