package azure

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/utils"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/kubesimplify/ksctl/api/resources"
)

// DelFirewall implements resources.CloudFactory.
func (obj *AzureProvider) DelFirewall(storage resources.StorageFactory) error {
	nsg := ""
	switch obj.metadata.role {
	case utils.ROLE_CP:
		nsg = azureCloudState.InfoControlPlanes.NetworkSecurityGroupName
	case utils.ROLE_WP:
		nsg = azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupName
	case utils.ROLE_LB:
		nsg = azureCloudState.InfoLoadBalancer.NetworkSecurityGroupName
	case utils.ROLE_DS:
		nsg = azureCloudState.InfoDatabase.NetworkSecurityGroupName
	default:
		return fmt.Errorf("invalid role")
	}
	if len(nsg) == 0 {
		storage.Logger().Success("[skip] firewall already deleted")
		return nil
	}

	pollerResponse, err := obj.client.BeginDeleteSecurityGrp(nsg, nil)
	if err != nil {
		return err
	}
	storage.Logger().Print("[azure] firewall deleting...", nsg)

	_, err = obj.client.PollUntilDoneDelNSG(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}
	switch obj.metadata.role {
	case utils.ROLE_CP:
		azureCloudState.InfoControlPlanes.NetworkSecurityGroupName = ""
		azureCloudState.InfoControlPlanes.NetworkSecurityGroupID = ""
	case utils.ROLE_WP:
		azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupID = ""
		azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupName = ""
	case utils.ROLE_LB:
		azureCloudState.InfoLoadBalancer.NetworkSecurityGroupID = ""
		azureCloudState.InfoLoadBalancer.NetworkSecurityGroupName = ""
	case utils.ROLE_DS:
		azureCloudState.InfoDatabase.NetworkSecurityGroupID = ""
		azureCloudState.InfoDatabase.NetworkSecurityGroupName = ""
	}

	if err := saveStateHelper(storage); err != nil {
		return err
	}
	storage.Logger().Success("[azure] Deleted network security group", nsg)

	return nil
}

// NewFirewall implements resources.CloudFactory.
func (obj *AzureProvider) NewFirewall(storage resources.StorageFactory) error {
	nsg := ""
	switch obj.metadata.role {
	case utils.ROLE_CP:
		nsg = azureCloudState.InfoControlPlanes.NetworkSecurityGroupName
	case utils.ROLE_WP:
		nsg = azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupName
	case utils.ROLE_LB:
		nsg = azureCloudState.InfoLoadBalancer.NetworkSecurityGroupName
	case utils.ROLE_DS:
		nsg = azureCloudState.InfoDatabase.NetworkSecurityGroupName
	default:
		return fmt.Errorf("invalid role")
	}
	if len(nsg) != 0 {
		storage.Logger().Success("[skip] firewall already created", nsg)
		return nil
	}

	var securityRules []*armnetwork.SecurityRule
	switch obj.metadata.role {
	case utils.ROLE_CP:
		securityRules = firewallRuleControlPlane()
	case utils.ROLE_WP:
		securityRules = firewallRuleWorkerPlane()
	case utils.ROLE_LB:
		securityRules = firewallRuleLoadBalancer()
	case utils.ROLE_DS:
		securityRules = firewallRuleDataStore()
	default:
		return fmt.Errorf("invalid role")
	}

	parameters := armnetwork.SecurityGroup{
		Location: to.Ptr(obj.region),
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: securityRules,
		},
	}

	pollerResponse, err := obj.client.BeginCreateSecurityGrp(obj.metadata.resName, parameters, nil)

	if err != nil {
		return err
	}
	switch obj.metadata.role {
	case utils.ROLE_CP:
		azureCloudState.InfoControlPlanes.NetworkSecurityGroupName = obj.metadata.resName
	case utils.ROLE_WP:
		azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupName = obj.metadata.resName
	case utils.ROLE_LB:
		azureCloudState.InfoLoadBalancer.NetworkSecurityGroupName = obj.metadata.resName
	case utils.ROLE_DS:
		azureCloudState.InfoDatabase.NetworkSecurityGroupName = obj.metadata.resName
	}

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	storage.Logger().Print("[azure] creating firewall...", obj.metadata.resName)

	resp, err := obj.client.PollUntilDoneCreateNSG(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}
	switch obj.metadata.role {
	case utils.ROLE_CP:
		azureCloudState.InfoControlPlanes.NetworkSecurityGroupID = *resp.ID
	case utils.ROLE_WP:
		azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupID = *resp.ID
	case utils.ROLE_LB:
		azureCloudState.InfoLoadBalancer.NetworkSecurityGroupID = *resp.ID
	case utils.ROLE_DS:
		azureCloudState.InfoDatabase.NetworkSecurityGroupID = *resp.ID
	}

	if err := saveStateHelper(storage); err != nil {
		return err
	}
	storage.Logger().Success("[azure] Created network security group", *resp.Name)

	return nil
}

// FIXME: add fine-grained rules
func firewallRuleControlPlane() (securityRules []*armnetwork.SecurityRule) {
	securityRules = append(securityRules, &armnetwork.SecurityRule{
		Name: to.Ptr("sample_inbound_6443"),
		Properties: &armnetwork.SecurityRulePropertiesFormat{
			SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
			SourcePortRange:          to.Ptr("*"),
			DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
			DestinationPortRange:     to.Ptr("*"),
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
			DestinationPortRange:     to.Ptr("*"),
			Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
			Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
			Priority:                 to.Ptr[int32](101),
			Description:              to.Ptr("sample network security group inbound port 30000-35000"),
			Direction:                to.Ptr(armnetwork.SecurityRuleDirectionOutbound),
		},
	})

	return
}

// FIXME: add fine-grained rules
func firewallRuleWorkerPlane() (securityRules []*armnetwork.SecurityRule) {
	securityRules = append(securityRules, &armnetwork.SecurityRule{
		Name: to.Ptr("sample_inbound_6443"),
		Properties: &armnetwork.SecurityRulePropertiesFormat{
			SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
			SourcePortRange:          to.Ptr("*"),
			DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
			DestinationPortRange:     to.Ptr("*"),
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
			DestinationPortRange:     to.Ptr("*"),
			Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
			Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
			Priority:                 to.Ptr[int32](101),
			Description:              to.Ptr("sample network security group inbound port 30000-35000"),
			Direction:                to.Ptr(armnetwork.SecurityRuleDirectionOutbound),
		},
	})
	return
}

// FIXME: add fine-grained rules
func firewallRuleLoadBalancer() (securityRules []*armnetwork.SecurityRule) {
	securityRules = []*armnetwork.SecurityRule{
		&armnetwork.SecurityRule{
			Name: to.Ptr("sample_inbound_6443"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
				SourcePortRange:          to.Ptr("*"),
				DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
				DestinationPortRange:     to.Ptr("*"),
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
				DestinationPortRange:     to.Ptr("*"),
				Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
				Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
				Priority:                 to.Ptr[int32](101),
				Description:              to.Ptr("sample network security group inbound port 30000-35000"),
				Direction:                to.Ptr(armnetwork.SecurityRuleDirectionOutbound),
			},
		},
	}
	return
}

// FIXME: add fine-grained rules
func firewallRuleDataStore() (securityRules []*armnetwork.SecurityRule) {
	securityRules = append(securityRules, &armnetwork.SecurityRule{
		Name: to.Ptr("sample_inbound_6443"),
		Properties: &armnetwork.SecurityRulePropertiesFormat{
			SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
			SourcePortRange:          to.Ptr("*"),
			DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
			DestinationPortRange:     to.Ptr("*"),
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
			DestinationPortRange:     to.Ptr("*"),
			Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
			Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
			Priority:                 to.Ptr[int32](101),
			Description:              to.Ptr("sample network security group inbound port 30000-35000"),
			Direction:                to.Ptr(armnetwork.SecurityRuleDirectionOutbound),
		},
	})
	return
}
