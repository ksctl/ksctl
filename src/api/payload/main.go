/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

package payload

import (
	"fmt"
	"os"
)

type Machine struct {
	Nodes int
	Cpu   string
	Mem   string
	Disk  string
}

type AwsProvider struct {
	ClusterName string
	HACluster   bool
	Region      string
	Spec        Machine
	AccessKey   string
	Secret      string
}

type AzureProvider struct {
	ClusterName         string
	HACluster           bool
	Region              string
	Spec                Machine
	SubscriptionID      string
	TenantID            string
	servicePrincipleKey string
	servicePrincipleID  string
}

type CivoProvider struct {
	ClusterName string
	APIKey      string
	HACluster   bool
	Region      string
	Spec        Machine
	Application string
	CNIPlugin   string
}

type LocalProvider struct {
	ClusterName string
	HACluster   bool
	Spec        Machine
}

//type Providers struct {
//	eks  *AwsProvider
//	aks  *AzureProvider
//	k3s  *CivoProvider
//	mk8s *LocalProvider
//}

// GetUserName returns current active username
func GetUserName() string {

	usrCmd := os.Getenv("HOME")

	//output, err := usrCmd.Output()
	//if err != nil {
	//	return ""
	//}
	//userName := strings.Trim(string(output), "\n")
	return usrCmd
}

type PrinterKubeconfigPATH interface {
	Printer(int)
}

var (
	CRED_PATH = fmt.Sprintf("%s/.ksctl/cred/", GetUserName())
)
