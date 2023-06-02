package azure

import (
	"context"
	"fmt"
	"strings"

	log "github.com/kubesimplify/ksctl/api/logger"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	util "github.com/kubesimplify/ksctl/api/utils"
)

func scriptWithoutCP_1(dbEndpoint, privateIPlb string) string {

	return fmt.Sprintf(`#!/bin/bash
cat <<EOF > control-setup.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="v1.24.6+k3s1" sh -s - server \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--datastore-endpoint "%s" \
	--tls-san %s
EOF

sudo chmod +x control-setup.sh
sudo ./control-setup.sh
`, dbEndpoint, privateIPlb)
}

func scriptWithCP_1() string {
	return `#!/bin/bash
sudo cat /var/lib/rancher/k3s/server/token
`
}

func scriptCP_n(dbEndpoint, privateIPlb, token string) string {
	return fmt.Sprintf(`#!/bin/bash
cat <<EOF > control-setupN.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="v1.24.6+k3s1" sh -s - server --token %s --datastore-endpoint="%s" --node-taint CriticalAddonsOnly=true:NoExecute --tls-san %s
EOF
log.Println
sudo chmod +x control-setupN.sh
sudo ./control-setupN.sh
`, token, dbEndpoint, privateIPlb)
}

func scriptKUBECONFIG() string {
	return `#!/bin/bash
sudo cat /etc/rancher/k3s/k3s.yaml`
}

// TODO: Add more firewallrules
func getControlPlaneFirewallRules() (securityRules []*armnetwork.SecurityRule) {
	securityRules = append(securityRules, &armnetwork.SecurityRule{
		Name: to.Ptr("sample_inbound_6443"),
		Properties: &armnetwork.SecurityRulePropertiesFormat{
			SourceAddressPrefix:      to.Ptr("*"),
			SourcePortRange:          to.Ptr("*"),
			DestinationAddressPrefix: to.Ptr("*"),
			DestinationPortRange:     to.Ptr("22"),
			Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
			Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
			Priority:                 to.Ptr[int32](100),
			Description:              to.Ptr("sample network security group inbound port 6443"),
			Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
		},
		// updated from herer  ----------------------------------------------------------
	}, &armnetwork.SecurityRule{
		Name: to.Ptr("deny_all_inbound"),
		Properties: &armnetwork.SecurityRulePropertiesFormat{
			SourceAddressPrefix:      to.Ptr("0.0.0.0/0"), // allow all source IP addresses
			SourcePortRange:          to.Ptr("*"),
			DestinationAddressPrefix: to.Ptr("*"),  // allow all destination IP addresses
			DestinationPortRange:     to.Ptr("22"), // allow inbound traffic only on port 22 (SSH)

			Protocol:    to.Ptr(armnetwork.SecurityRuleProtocolTCP),
			Access:      to.Ptr(armnetwork.SecurityRuleAccessDeny),
			Priority:    to.Ptr[int32](112),
			Description: to.Ptr("expose port 22 for controlplane from public -->"),
			Direction:   to.Ptr(armnetwork.SecurityRuleDirectionInbound),
		},
	}, &armnetwork.SecurityRule{ // ignore this going to do it later
		Name: to.Ptr("outbound_secure_communication"),
		Properties: &armnetwork.SecurityRulePropertiesFormat{
			SourceAddressPrefix:      to.Ptr("define here"),
			SourcePortRange:          to.Ptr("*"),
			DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
			// DestinationPortRange:     to.Ptr("443"), // allow outbound traffic on port 443 (HTTPS)
			Protocol:    to.Ptr(armnetwork.SecurityRuleProtocolTCP),
			Access:      to.Ptr(armnetwork.SecurityRuleAccessAllow),
			Priority:    to.Ptr[int32](200),
			Description: to.Ptr("Allow outbound secure communication"),
			Direction:   to.Ptr(armnetwork.SecurityRuleDirectionOutbound),
		},
	})
	return
}

func (obj *AzureProvider) FetchKUBECONFIG(logging log.Logger, publicIP string) (string, error) {
	obj.SSH_Payload.PublicIP = publicIP
	obj.SSH_Payload.Output = ""
	err := obj.SSH_Payload.SSHExecute(logging, util.EXEC_WITH_OUTPUT, scriptKUBECONFIG(), true)

	if err != nil {
		return "", nil
	}

	return obj.SSH_Payload.Output, nil
}

// GetTokenFromCP_1 used to extract the K3S_TOKEN from the first Controlplane node
func (obj *AzureProvider) GetTokenFromCP_1(logger log.Logger, PublicIP string) string {
	obj.SSH_Payload.PublicIP = PublicIP
	obj.SSH_Payload.Output = ""
	err := obj.SSH_Payload.SSHExecute(logger, util.EXEC_WITH_OUTPUT, scriptWithCP_1(), true)
	if err != nil {
		return ""
	}
	token := obj.SSH_Payload.Output
	obj.SSH_Payload.Output = ""
	token = strings.Trim(token, "\n")
	obj.Config.K3sToken = token

	obj.ConfigWriter(logger, "ha")

	return token
}

// HelperExecNoOutputControlPlane helps with script execution without returning us the output
func (obj *AzureProvider) HelperExecNoOutputControlPlane(logger log.Logger, publicIP, script string, fastMode bool) error {
	obj.SSH_Payload.PublicIP = publicIP
	obj.SSH_Payload.Output = ""
	err := obj.SSH_Payload.SSHExecute(logger, util.EXEC_WITH_OUTPUT, script, fastMode)
	if err != nil {
		return err
	}

	return nil
}

// HelperExecOutputControlPlane helps with script execution and also returns the script output
func (obj *AzureProvider) HelperExecOutputControlPlane(logger log.Logger, publicIP, script string, fastMode bool) (string, error) {
	obj.SSH_Payload.Output = ""
	obj.SSH_Payload.PublicIP = publicIP
	err := obj.SSH_Payload.SSHExecute(logger, util.EXEC_WITH_OUTPUT, script, fastMode)
	if err != nil {
		return "", err
	}
	return obj.SSH_Payload.Output, nil
}

func (obj *AzureProvider) createControlPlane(ctx context.Context, logger log.Logger, indexOfNode int) error {
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

	vmName := fmt.Sprintf("%s-cp-%d", obj.ClusterName, indexOfNode)

	publicIP, err := obj.CreatePublicIP(ctx, logger, vmName+"-pub-ip")
	if err != nil {
		return err
	}
	obj.Config.InfoControlPlanes.PublicIPNames = append(obj.Config.InfoControlPlanes.PublicIPNames, *publicIP.Name)

	// network security group
	if len(obj.Config.InfoControlPlanes.NetworkSecurityGroupName) == 0 {
		nsg, err := obj.CreateNSG(ctx, logger, obj.ClusterName+"-cp-nsg", getControlPlaneFirewallRules())
		if err != nil {
			return err
		}

		obj.Config.InfoControlPlanes.NetworkSecurityGroupName = *nsg.Name
		obj.Config.InfoControlPlanes.NetworkSecurityGroupID = *nsg.ID
	}

	networkInterface, err := obj.CreateNetworkInterface(ctx, logger, obj.Config.ResourceGroupName, vmName+"-nic", obj.Config.SubnetID, *publicIP.ID, obj.Config.InfoControlPlanes.NetworkSecurityGroupID)
	if err != nil {
		return err
	}
	obj.Config.InfoControlPlanes.NetworkInterfaceNames = append(obj.Config.InfoControlPlanes.NetworkInterfaceNames, *networkInterface.Name)

	obj.Config.InfoControlPlanes.Names = append(obj.Config.InfoControlPlanes.Names, vmName)
	obj.Config.InfoControlPlanes.DiskNames = append(obj.Config.InfoControlPlanes.DiskNames, vmName+"-disk")

	_, err = obj.CreateVM(ctx, logger, vmName, *networkInterface.ID, vmName+"-disk", "")
	if err != nil {
		return err
	}
	obj.Config.InfoControlPlanes.PublicIPs = append(obj.Config.InfoControlPlanes.PublicIPs, *publicIP.Properties.IPAddress)
	obj.Config.InfoControlPlanes.PrivateIPs = append(obj.Config.InfoControlPlanes.PrivateIPs, *networkInterface.Properties.IPConfigurations[0].Properties.PrivateIPAddress)
	logger.Info("💻 Booted Control plane VM", vmName)
	return nil
}
