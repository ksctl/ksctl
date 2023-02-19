package azure

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

func scriptWP(privateIPlb, token string) string {
	return fmt.Sprintf(`#!/bin/bash
export SECRET='%s'
curl -sfL https://get.k3s.io | sh -s - agent --token=$SECRET --server https://%s:6443
`, token, privateIPlb)
}

func getWorkerPlaneFirewallRules() (securityRules []*armnetwork.SecurityRule) {
	securityRules = append(securityRules, &armnetwork.SecurityRule{
		Name: to.Ptr("sample_inbound all open"),
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
		Name: to.Ptr("sample_inbound all open"),
		Properties: &armnetwork.SecurityRulePropertiesFormat{
			SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
			SourcePortRange:          to.Ptr("*"),
			DestinationAddressPrefix: to.Ptr("10.1.0.0/16"),
			DestinationPortRange:     to.Ptr("3306"),
			Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
			Access:                   to.Ptr(armnetwork.SecurityRuleAccessDeny),
			Priority:                 to.Ptr[int32](100),
			Description:              to.Ptr("sample network security group outbound port 3306"),
			Direction:                to.Ptr(armnetwork.SecurityRuleDirectionOutbound),
		},
	})
	return
}
