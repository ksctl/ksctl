package eks

import "fmt"

type Aws interface {
	CreteBucket(bucketName string) error
	DeleteFile(bucketName string, fileName string) error
	DeleteBucket(bucketName string) error
	CreateCluster(clusterName string, nodeType string, nodeCount int64) error
	DeleteCluster(clusterName string) error
	StopCluster(clusterName string) error
	StartCluster(clusterName string) error
	ClsuterStatus(clusterName string) (string, error)
	BucketType() string
	Create_VPC() (string, error)
}

type Cluster struct {
	ClusterName     string
	NodeType        string
	NodeCount       int64
	Vpc             string
	Subnet          string
	InternetGateway string
	RouteTable      string
	SecurityGroup   string
}

func (a *AWSCLUSTER) CreateCluster(clusterName string, nodeType string, nodeCount int64) error {
	cred := Credentials()
	if !cred {
		return fmt.Errorf("Credentials not found \n make sure to add your aws credentials")
	}

	a.Create_VPC()

	return nil
}

func (a *AWSCLUSTER) DeleteCluster(clusterName string) error {

	return nil
}

func (a *AWSCLUSTER) StopCluster(clusterName string) error {

	return nil
}

func (a *AWSCLUSTER) StartCluster(clusterName string) error {

	return nil
}

func (a *AWSCLUSTER) ClsuterStatus(clusterName string) (string, error) {

	a

	return "", nil
}

func (a *AWSCLUSTER) Create_VPC() (string, error) {

	// create  a vpc of cluster name
	// create a subnet of cluster name
	// create a internet gateway of cluster name
	// create a route table of cluster name
	// create a security group of cluster name
	// create a key pair of cluster name
	// create a ec2 instance of cluster name
	// create a load balancer of cluster name
	// create a target group of cluster name

	return "", nil
}

func (a *AWSCLUSTER) BucketType() string {

	return "aws"
}

func (a *AWSCLUSTER) CreteBucket(bucketName string) error {

	return nil
}

func (a *AWSCLUSTER) ListBuckets() ([]string, error) {

	return nil, nil
}

func (a *AWSCLUSTER) UploadFile(bucketName string, fileName string, fileContent []byte) error {

	return nil
}

func (a *AWSCLUSTER) DownloadFile(bucketName string, fileName string) ([]byte, error) {

	return nil, nil
}

func (a *AWSCLUSTER) DeleteFile(bucketName string, fileName string) error {

	return nil
}

func (a *AWSCLUSTER) DeleteBucket(bucketName string) error {

	return nil
}

func (a *AWSCLUSTER) Credentials() bool {

	return true
}
