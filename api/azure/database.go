package azure

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

func generateDBPassword(passwordLen int) string {
	var password strings.Builder
	var (
		lowerCharSet = "abcdedfghijklmnopqrst"
		upperCharSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numberSet    = "0123456789"
		allCharSet   = lowerCharSet + upperCharSet + numberSet
	)
	rand.Seed(time.Now().Unix())

	for i := 0; i < passwordLen; i++ {
		random := rand.Intn(len(allCharSet))
		password.WriteString(string(allCharSet[random]))
	}

	inRune := []rune(password.String())
	rand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})

	return string(inRune)
}

func scriptDB(password string) string {
	return fmt.Sprintf(`#!/bin/bash
sudo apt update
sudo apt install -y mysql-server

sudo systemctl start mysql

sudo systemctl enable mysql

cat <<EOF > mysqld.cnf
[mysqld]
user		= mysql
bind-address		= 0.0.0.0
#mysqlx-bind-address	= 127.0.0.1

key_buffer_size		= 16M
myisam-recover-options  = BACKUP
log_error = /var/log/mysql/error.log
max_binlog_size   = 100M

EOF

sudo mv mysqld.cnf /etc/mysql/mysql.conf.d/mysqld.cnf

sudo systemctl restart mysql

sudo mysql -e "create user 'ksctl' identified by '%s';"
sudo mysql -e "create database ksctldb; grant all on ksctldb.* to 'ksctl';"

`, password)
}

func getDatabaseFirewallRules() (securityRules []*armnetwork.SecurityRule) {
	securityRules = append(securityRules, &armnetwork.SecurityRule{
		Name: to.Ptr("sample_inbound_all"),
		Properties: &armnetwork.SecurityRulePropertiesFormat{
			SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
			SourcePortRange:          to.Ptr("*"),
			DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
			DestinationPortRange:     to.Ptr("*"),
			// DestinationPortRange:     to.Ptr("3306"),
			Protocol:    to.Ptr(armnetwork.SecurityRuleProtocolTCP),
			Access:      to.Ptr(armnetwork.SecurityRuleAccessAllow),
			Priority:    to.Ptr[int32](100),
			Description: to.Ptr("sample network security group inbound port 3306"),
			Direction:   to.Ptr(armnetwork.SecurityRuleDirectionInbound),
		},
	},
		&armnetwork.SecurityRule{
			Name: to.Ptr("sample_outbound_port_all"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
				SourcePortRange:          to.Ptr("*"),
				DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
				DestinationPortRange:     to.Ptr("*"),
				Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
				Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
				Priority:                 to.Ptr[int32](101),
				Description:              to.Ptr("sample network security group outbound port All"),
				Direction:                to.Ptr(armnetwork.SecurityRuleDirectionOutbound),
			},
		})
	return
}

func (obj *AzureProvider) createDatabase(ctx context.Context) error {
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
	generatedPassword := generateDBPassword(20)

	publicIP, err := obj.CreatePublicIP(ctx, obj.ClusterName+"-db-pub-ip")
	if err != nil {
		return err
	}
	obj.Config.InfoDatabase.PublicIPName = *publicIP.Name

	// network security group
	if len(obj.Config.InfoDatabase.NetworkSecurityGroupName) == 0 {
		nsg, err := obj.CreateNSG(ctx, obj.ClusterName+"-db-nsg", getDatabaseFirewallRules())
		if err != nil {
			return err
		}

		obj.Config.InfoDatabase.NetworkSecurityGroupName = *nsg.Name
		obj.Config.InfoDatabase.NetworkSecurityGroupID = *nsg.ID
	}

	networkInterface, err := obj.CreateNetworkInterface(ctx, obj.Config.ResourceGroupName, obj.ClusterName+"-db-nic", obj.Config.SubnetID, *publicIP.ID, obj.Config.InfoDatabase.NetworkSecurityGroupID)
	if err != nil {
		return err
	}
	obj.Config.InfoDatabase.NetworkInterfaceName = *networkInterface.Name

	obj.Config.InfoDatabase.Name = obj.ClusterName + "-db"
	obj.Config.InfoDatabase.DiskName = obj.ClusterName + "-db-disk"

	obj.Config.InfoDatabase.PrivateIP = *networkInterface.Properties.IPConfigurations[0].Properties.PrivateIPAddress
	_, err = obj.CreateVM(ctx, obj.ClusterName+"-db", *networkInterface.ID, obj.ClusterName+"-db-disk", scriptDB(generatedPassword))
	if err != nil {
		return err
	}

	obj.Config.InfoDatabase.PublicIP = *publicIP.Properties.IPAddress

	obj.Config.DBEndpoint = fmt.Sprintf("mysql://ksctl:%s@tcp(%s:3306)/ksctldb", generatedPassword, obj.Config.InfoDatabase.PrivateIP)
	log.Println("💻 Booted Database VM ")
	return nil
}
