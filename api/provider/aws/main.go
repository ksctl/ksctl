package aws

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/kubesimplify/ksctl/api/resources"
)

var (
	awsCloudState  *StateConfiguration
	clusterDirName string
	clusterType    string
	ctx            context.Context
)

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
	Vpc           string `json:"vpc"`
	Name          string `json:"name"`
	DiskSize      string `json:"disk_size"`
	InstanceType  string `json:"instance_type"`
	Subnet        string `json:"subnet"`
	SecurityGroup string `json:"security_group"`
	PublicIP      string `json:"public_ip"`
	PrivateIP     string `json:"private_ip"`
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

	InfoControlPlanes AWSStateVms `json:"info_control_planes"`
	InfoWorkerPlanes  AWSStateVms `json:"info_worker_planes"`
	InfoDatabase      AWSStateVms `json:"info_database"`
	InfoLoadBalancer  AWSStateVms `json:"info_load_balancer"`

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
		Region:      metadata.Region,
		HACluster:   metadata.IsHA,
		Metadata: Metadata{
			K8sVersion: metadata.K8sVersion,
			K8sName:    metadata.K8sDistro,
		},
	}, nil
}

func (obj *AwsProvider) Name(resName string) resources.CloudFactory {
	obj.Metadata.ResName = resName
	return nil
}

func (obj *AwsProvider) NEWCLIENT() (*session.Session, error) {

	NewSession, err := session.NewSession(&aws.Config{
		Region:      aws.String(AWS_REGION),
		Credentials: credentials.NewStaticCredentials(obj.AccessKeyID, obj.Secret, ""),
	})
	if err != nil {
		log.Fatal(err)
	}
	return NewSession, nil
}
