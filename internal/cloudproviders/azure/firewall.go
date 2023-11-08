package azure

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils/consts"
)

// DelFirewall implements resources.CloudFactory.
func (obj *AzureProvider) DelFirewall(storage resources.StorageFactory) error {
	role := obj.metadata.role
	obj.mxRole.Unlock()

	log.Debug("Printing", "role", role)

	nsg := ""
	switch role {
	case consts.RoleCp:
		nsg = azureCloudState.InfoControlPlanes.NetworkSecurityGroupName
	case consts.RoleWp:
		nsg = azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupName
	case consts.RoleLb:
		nsg = azureCloudState.InfoLoadBalancer.NetworkSecurityGroupName
	case consts.RoleDs:
		nsg = azureCloudState.InfoDatabase.NetworkSecurityGroupName
	default:
		return fmt.Errorf("invalid role")
	}

	if len(nsg) == 0 {
		log.Print("skipped firewall already deleted")
		return nil
	}

	pollerResponse, err := obj.client.BeginDeleteSecurityGrp(nsg, nil)
	if err != nil {
		return log.NewError(err.Error())
	}
	log.Print("firewall deleting...", "name", nsg)

	_, err = obj.client.PollUntilDoneDelNSG(ctx, pollerResponse, nil)
	if err != nil {
		return log.NewError(err.Error())
	}
	switch role {
	case consts.RoleCp:
		azureCloudState.InfoControlPlanes.NetworkSecurityGroupName = ""
		azureCloudState.InfoControlPlanes.NetworkSecurityGroupID = ""
	case consts.RoleWp:
		azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupID = ""
		azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupName = ""
	case consts.RoleLb:
		azureCloudState.InfoLoadBalancer.NetworkSecurityGroupID = ""
		azureCloudState.InfoLoadBalancer.NetworkSecurityGroupName = ""
	case consts.RoleDs:
		azureCloudState.InfoDatabase.NetworkSecurityGroupID = ""
		azureCloudState.InfoDatabase.NetworkSecurityGroupName = ""
	}

	if err := saveStateHelper(storage); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("Deleted network security group", "name", nsg)

	return nil
}

// NewFirewall implements resources.CloudFactory.
func (obj *AzureProvider) NewFirewall(storage resources.StorageFactory) error {
	name := obj.metadata.resName
	role := obj.metadata.role
	obj.mxRole.Unlock()
	obj.mxName.Unlock()

	log.Debug("Printing", "name", name, "role", role)

	nsg := ""
	switch role {
	case consts.RoleCp:
		nsg = azureCloudState.InfoControlPlanes.NetworkSecurityGroupName
	case consts.RoleWp:
		nsg = azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupName
	case consts.RoleLb:
		nsg = azureCloudState.InfoLoadBalancer.NetworkSecurityGroupName
	case consts.RoleDs:
		nsg = azureCloudState.InfoDatabase.NetworkSecurityGroupName
	default:
		return log.NewError("invalid role")
	}
	if len(nsg) != 0 {
		log.Success("skipped firewall already created", "name", nsg)
		return nil
	}

	var securityRules []*armnetwork.SecurityRule
	switch role {
	case consts.RoleCp:
		securityRules = firewallRuleControlPlane()
	case consts.RoleWp:
		securityRules = firewallRuleWorkerPlane()
	case consts.RoleLb:
		securityRules = firewallRuleLoadBalancer()
	case consts.RoleDs:
		securityRules = firewallRuleDataStore()
	default:
		return log.NewError("invalid role")
	}

	log.Debug("Printing", "firewallrule", securityRules)

	parameters := armnetwork.SecurityGroup{
		Location: to.Ptr(obj.region),
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: securityRules,
		},
	}

	pollerResponse, err := obj.client.BeginCreateSecurityGrp(name, parameters, nil)

	if err != nil {
		return log.NewError(err.Error())
	}
	switch role {
	case consts.RoleCp:
		azureCloudState.InfoControlPlanes.NetworkSecurityGroupName = name
	case consts.RoleWp:
		azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupName = name
	case consts.RoleLb:
		azureCloudState.InfoLoadBalancer.NetworkSecurityGroupName = name
	case consts.RoleDs:
		azureCloudState.InfoDatabase.NetworkSecurityGroupName = name
	}

	if err := saveStateHelper(storage); err != nil {
		return log.NewError(err.Error())
	}

	log.Print("creating firewall...", "name", name)

	resp, err := obj.client.PollUntilDoneCreateNSG(ctx, pollerResponse, nil)
	if err != nil {
		return log.NewError(err.Error())
	}
	switch role {
	case consts.RoleCp:
		azureCloudState.InfoControlPlanes.NetworkSecurityGroupID = *resp.ID
	case consts.RoleWp:
		azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupID = *resp.ID
	case consts.RoleLb:
		azureCloudState.InfoLoadBalancer.NetworkSecurityGroupID = *resp.ID
	case consts.RoleDs:
		azureCloudState.InfoDatabase.NetworkSecurityGroupID = *resp.ID
	}

	if err := saveStateHelper(storage); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("Created network security group", "name", *resp.Name)

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
