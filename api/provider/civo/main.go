package civo

import (
	"errors"
	"fmt"
	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
	"github.com/kubesimplify/ksctl/api/utils"
	"os"
)

type InstanceID struct {
	ControlNodes     []string `json:"controlnodeids"`
	WorkerNodes      []string `json:"workernodeids"`
	LoadBalancerNode []string `json:"loadbalancernodeids"`
	DatabaseNode     []string `json:"databasenodeids"`
}

type NetworkID struct {
	FirewallIDControlPlaneNode string `json:"fwidcontrolplanenode"`
	FirewallIDWorkerNode       string `json:"fwidworkernode"`
	FirewallIDLoadBalancerNode string `json:"fwidloadbalancenode"`
	FirewallIDDatabaseNode     string `json:"fwiddatabasenode"`
	NetworkID                  string `json:"clusternetworkid"`
}

type InstanceIP struct {
	IPControlplane        []string
	IPWorkerPlane         []string
	IPLoadbalancer        string
	IPDataStore           []string
	PrivateIPControlplane []string
	PrivateIPWorkerPlane  []string
	PrivateIPLoadbalancer string
	PrivateIPDataStore    []string
}

type StateConfiguration struct {
	ClusterName string     `json:"clustername"`
	Region      string     `json:"region"`
	SSHID       string     `json:"ssh_id"`
	InstanceIDs InstanceID `json:"instanceids"`
	NetworkIDs  NetworkID  `json:"networkids"`
	IPv4        InstanceIP `json:"ipv4_addr"`
}

var (
	civoCloudState *StateConfiguration
	civoClient     *civogo.Client
	clusterDirName string
	clusterType    string // it stores the ha or managed
)

const (
	FILE_PERM_CLUSTER_DIR        = os.FileMode(0750)
	FILE_PERM_CLUSTER_STATE      = os.FileMode(0640)
	FILE_PERM_CLUSTER_KUBECONFIG = os.FileMode(0750)
	STATE_FILE_NAME              = string("cloud-state.json")
)

type Metadata struct {
	ResName string
	Role    string
	VmType  string
	Public  bool
}

type CivoProvider struct {
	ClusterName string `json:"cluster_name"`
	APIKey      string `json:"api_key"`
	HACluster   bool   `json:"ha_cluster"`
	Region      string `json:"region"`
	Application string `json:"application"`
	CNIPlugin   string `json:"cni_plugin"`
	SSHPath     string `json:"ssh_key"` // do check what need to be here
	Metadata
}

// GetStateForHACluster implements resources.CloudInfrastructure.
func (client *CivoProvider) GetStateForHACluster(state resources.StateManagementInfrastructure) (cloud.CloudResourceState, error) {
	payload := cloud.CloudResourceState{
		SSHState:          cloud.SSHPayload{PathPrivateKey: "abcd/rdcewcf"},
		Metadata:          cloud.Metadata{ClusterName: client.ClusterName},
		IPv4ControlPlanes: civoCloudState.InstanceIDs.ControlNodes,
	}
	return payload, nil
}

func (obj *CivoProvider) InitState(state resources.StateManagementInfrastructure, operation string) error {
	clusterDirName = obj.ClusterName + " " + obj.Region
	if obj.HACluster {
		clusterType = "ha"
	} else {
		clusterType = "managed"
	}

	var err error

	switch operation {
	case "create":
		if civoCloudState != nil {
			return errors.New("[FATAL] already initialized")
		}
		civoCloudState = &StateConfiguration{}
	case "delete":
		// fetch from the state from store
		// TODO: add the fetch expisting state (if failed return error)
		err := loadStateHelper(state, generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, STATE_FILE_NAME))
		if err != nil {
			return err
		}
	default:
		return errors.New("Invalid operation for init state")
	}

	civoClient, err = civogo.NewClient(fetchAPIKey(), obj.Region)
	if err != nil {
		return err
	}

	fmt.Println("[civo] Civo cloud state", civoCloudState)
	return nil
}

func ReturnCivoStruct(metadata resources.Metadata) (*CivoProvider, error) {
	if err := validationOfArguments(metadata.ClusterName, metadata.Region); err != nil {
		return nil, err
	}
	return &CivoProvider{
		ClusterName: metadata.ClusterName,
		Region:      metadata.Region,
		HACluster:   metadata.IsHA,
	}, nil
}

// it will contain the name of the resource to be created
func (cloud *CivoProvider) Name(resName string) resources.CloudInfrastructure {
	cloud.Metadata.ResName = resName
	return cloud
}

// it will contain whether the resource to be created belongs for controlplane component or loadbalancer...
func (cloud *CivoProvider) Role(resRole string) resources.CloudInfrastructure {
	cloud.Metadata.Role = resRole
	return cloud
}

// it will contain which vmType to create
func (cloud *CivoProvider) VMType(size string) resources.CloudInfrastructure {
	cloud.Metadata.VmType = size
	return cloud
}

// whether to have the resource as public or private (i.e. VMs)
func (cloud *CivoProvider) Visibility(toBePublic bool) resources.CloudInfrastructure {
	cloud.Metadata.Public = toBePublic
	return cloud
}
