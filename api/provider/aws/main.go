package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/kubesimplify/ksctl/api/resources"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// var (
// 	awsCloudState  *StateConfiguration
// 	clusterDirName string
// 	clusterType    string
// 	ctx            context.Context
// )

type Credential struct {
	AccessKeyID string `json:"access_key_id"`
	Secret      string `json:"secret_access_key"`
}

type AwsProvider struct {
	ClusterName string `json:"cluster_name"`
	HACluster   bool   `json:"ha_cluster"`
	Region      string `json:"region"`
	VPC         string `json:"vpc"`
	AccessKeyID string `json:"access_key_id"`
	Secret      string `json:"secret_access_key"`
	Session     *session.Session
	Metadata    Metadata

	SSHPath string `json:"ssh_key"`
}
type AWSStateVms struct {
	Names                 []string `json:"names"`
	SecurityGroupName     string   `json:"network_security_group_name"`
	SecurityGroupID       string   `json:"network_security_group_id"`
	DiskNames             []string `json:"disk_name"`
	PublicIPNames         []string `json:"public_ip_"`
	PrivateIPs            []string `json:"private_ip"`
	PublicIPs             []string `json:"public_ip"`
	NetworkInterfaceNames []string `json:"network_interface_name"`
	NetworkInterfaceIDs   []string `json:"network_interface_id"`
	Hostnames             []string `json:"hostname"`
	PublicIPIDs           []string `json:"publicipids"`
}

var (
	azureCloudState *StateConfiguration

	clusterDirName string
	clusterType    string // it stores the ha or managed

)

const (
	FILE_PERM_CLUSTER_DIR        = os.FileMode(0750)
	FILE_PERM_CLUSTER_STATE      = os.FileMode(0640)
	FILE_PERM_CLUSTER_KUBECONFIG = os.FileMode(0755)
	STATE_FILE_NAME              = string("cloud-state.json")
	KUBECONFIG_FILE_NAME         = string("kubeconfig")
)

type AWSStateVm struct {
	Vpc           string `json:"vpc"`
	Name          string `json:"name"`
	DiskSize      string `json:"disk_size"`
	InstanceType  string `json:"instance_type"`
	Subnet        string `json:"subnet"`
	SecurityGroup string `json:"security_group"`
	PublicIPName
	PublicIP  string `json:"public_ip"`
	PrivateIP string `json:"private_ip"`
}

type StateConfiguration struct {
	IsCompleted        bool
	ClusterName        string `json:"cluster_name"`
	Region             string `json:"region"`
	VPC                string `json:"vpc"`
	SSHUSER            string `json:"ssh_user"`
	SSHPrivateKeyLoc   string `json:"ssh_private_key_loc"`
	SSHKeyName         string `json:"ssh_key_name"`
	ManagedClusterName string `json:"managed_cluster_name"`
	NoManagedNodes     int    `json:"no_managed_nodes"`
	SubnetName         string `json:"subnet_name"`
	SubnetID           string `json:"subnet_id"`
	SecurityGroupName  string `json:"security_group_name"`
	SecurityGroupID    string `json:"security_group_id"`
	GatewayName        string `json:"gateway_name"`
	GatewayID          string `json:"gateway_id"`
	RouteTableName     string `json:"route_table_name"`
	RouteTableID       string `json:"route_table_id"`

	InfoControlPlanes AWSStateVms `json:"info_control_planes"`
	InfoWorkerPlanes  AWSStateVms `json:"info_worker_planes"`
	InfoDatabase      AWSStateVms `json:"info_database"`
	InfoLoadBalancer  AWSStateVm  `json:"info_load_balancer"`

	KubernetesDistro string `json:"k8s_distro"`
	KubernetesVer    string `json:"k8s_version"`
}

type Metadata struct {
	ResName string
	Role    string
	VmType  string
	Public  bool

	Apps    string
	Cni     string
	Version string

	NoCP int
	NoWP int
	NoDS int

	K8sName    string
	K8sVersion string
}

func ReturnAwsStruct(metadata resources.Metadata) (*AwsProvider, error) {
	return &AwsProvider{
		ClusterName: metadata.ClusterName,
		Region:      "ap-south-1",
		HACluster:   metadata.IsHA,
		Metadata: Metadata{
			K8sVersion: metadata.K8sVersion,
			K8sName:    metadata.K8sDistro,
		},
		AccessKeyID: "",
		Secret:      "",
	}, nil
}

func (obj *AwsProvider) Name(resName string) resources.CloudFactory {
	obj.Metadata.ResName = resName
	return nil
}

func (obj *AwsProvider) NEWCLIENT() (aws.Config, error) {

	// NewSession, err := session.NewSession(&aws.Config{
	// 	Region:      aws.String(obj.Region),
	// 	Credentials: credentials.NewStaticCredentials(obj.AccessKeyID, obj.Secret, ""),
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }

	NewSession, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("ap-south-1"),
		config.WithSharedConfigProfile("default"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("add key", "add token", "")),
	)
	if err != nil {
		panic(err)
	}

	NewSession.Region = obj.Region

	fmt.Println("AWS Session Created Successfully")

	return NewSession, nil

}
