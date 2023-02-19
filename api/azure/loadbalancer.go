package azure

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

func scriptLB() string {
	return `#!/bin/bash
sudo apt update
sudo apt install haproxy -y
sudo systemctl start haproxy && sudo systemctl enable haproxy
`
}

func configLBscript(controlPlaneIPs []string) string {
	script := `sudo cat <<EOF > /etc/haproxy/haproxy.cfg
frontend kubernetes-frontend
  bind *:6443
  mode tcp
  option tcplog
  timeout client 10s
  default_backend kubernetes-backend

backend kubernetes-backend
  timeout connect 10s
  timeout server 10s
  mode tcp
  option tcp-check
  balance roundrobin
`

	for index, controlPlaneIP := range controlPlaneIPs {
		script += fmt.Sprintf(`  server k3sserver-%d %s check
`, index+1, controlPlaneIP)
	}

	script += `EOF

sudo systemctl restart haproxy
`
	return script
}

func getLoadBalancerFirewallRules() (securityRules []*armnetwork.SecurityRule) {
	securityRules = append(securityRules, &armnetwork.SecurityRule{
		Name: to.Ptr("sample_inbound_6443"),
		Properties: &armnetwork.SecurityRulePropertiesFormat{
			SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
			SourcePortRange:          to.Ptr("*"),
			DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
			DestinationPortRange:     to.Ptr("6443"),
			Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
			Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
			Priority:                 to.Ptr[int32](100),
			Description:              to.Ptr("sample network security group inbound port 6443"),
			Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
		},
	},
	)
	// &armnetwork.SecurityRule{
	// 	Name: to.Ptr("sample_outbound_6443"),
	// 	Properties: &armnetwork.SecurityRulePropertiesFormat{
	// 		SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
	// 		SourcePortRange:          to.Ptr("*"),
	// 		DestinationAddressPrefix: to.Ptr("10.1.0.0/16"),
	// 		DestinationPortRange:     to.Ptr("6443"),
	// 		Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
	// 		Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
	// 		Priority:                 to.Ptr[int32](101),
	// 		Description:              to.Ptr("sample network security group outbound port 6443"),
	// 		Direction:                to.Ptr(armnetwork.SecurityRuleDirectionOutbound),
	// 	},
	// })
	return
}

func (obj *AzureProvider) createLoadBalancer(ctx context.Context) error {
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

	publicIP, err := obj.CreatePublicIP(ctx, obj.ClusterName+"-lb-pub-ip")
	if err != nil {
		return err
	}
	obj.Config.InfoLoadBalancer.PublicIPName = *publicIP.Name
	// TODO: call config writer

	// network security group
	if len(obj.Config.InfoLoadBalancer.NetworkSecurityGroupName) == 0 {
		nsg, err := obj.CreateNSG(ctx, obj.ClusterName+"-lb-nsg", getLoadBalancerFirewallRules())
		if err != nil {
			return err
		}

		obj.Config.InfoLoadBalancer.NetworkSecurityGroupName = *nsg.Name
		obj.Config.InfoLoadBalancer.NetworkSecurityGroupID = *nsg.ID
	}

	networkInterface, err := obj.CreateNetworkInterface(ctx, obj.Config.ResourceGroupName, obj.ClusterName+"-lb-nic", obj.Config.SubnetID, *publicIP.ID, obj.Config.InfoLoadBalancer.NetworkSecurityGroupID)
	if err != nil {
		return err
	}
	obj.Config.InfoLoadBalancer.NetworkInterfaceName = *networkInterface.Name

	obj.Config.InfoLoadBalancer.Name = obj.ClusterName + "-lb"
	obj.Config.InfoLoadBalancer.DiskName = obj.ClusterName + "-lb-disk"

	_, err = obj.CreateVM(ctx, obj.ClusterName+"-lb", *networkInterface.ID, obj.ClusterName+"-lb-disk", scriptLB())
	if err != nil {
		return err
	}
	log.Println("💻 Booted LoadBalancer VM ")
	return nil
}
