package azure

import (
	"fmt"

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
		Name: to.Ptr("sample_inbound 30-35k"),
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
