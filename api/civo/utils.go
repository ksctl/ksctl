/*
Kubesimplify
Credit to @civo
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

package civo

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/kubesimplify/ksctl/api/logger"

	"github.com/civo/civogo"
	util "github.com/kubesimplify/ksctl/api/utils"
)

// NOTE: where are the configs stored
// .ksctl
// |--- config
// |    |--- civo
// .    .    |--- managed {contains (config, info.json)}
// .    .    |--- ha {contains (config, info.json, keypair, keypair.pub)}
//

const (
	SSH_PAUSE_IN_SECONDS = 20
	MAX_RETRY_COUNT      = 8
)

type LoadBalancerRet struct {
	PublicIP   string
	InstanceID *civogo.Instance
}

type HACollection interface {
	DeleteInstances(log.Logger) error
	DeleteNetworks(log.Logger) error

	DeleteInstance(string) error
	DeleteFirewall(string) error
	DeleteNetwork(string) error

	GetNetwork(string) (*civogo.Network, error)
	GetInstance(string) (*civogo.Instance, error)

	CreateFirewall(string) (*civogo.FirewallResult, error)
	CreateNetwork(log.Logger, string) error
	CreateInstance(string, string, string, string, bool) (*civogo.Instance, error)

	SaveKubeconfig(log.Logger, string) error

	CreateLoadbalancer(log.Logger) (*civogo.Instance, error)
	CreateControlPlane(log.Logger, int) (*civogo.Instance, error)
	CreateWorkerNode(log.Logger, int, string, string) (*civogo.Instance, error)
	CreateDatabase(log.Logger) (string, error)
	GetTokenFromCP_1(log.Logger, *civogo.Instance) string

	UploadSSHKey(log.Logger) error
	CreateSSHKeyPair(log.Logger, string) error
	DeleteSSHKeyPair() error
	ConfigLoadBalancer(log.Logger, *civogo.Instance, []string) error
	FetchKUBECONFIG(*civogo.Instance) (string, error)
	HelperExecNoOutputControlPlane(string, string, bool) error
	HelperExecOutputControlPlane(string, string, bool) (string, error)
}

type HAType struct {
	Client        *civogo.Client // CIVO client obj
	ClusterName   string         // clusterName provided by the user
	DiskImgID     string         // disk Img ID for ubuntu
	NetworkID     string         // network id
	NodeSize      string         // e.x. g3.medium
	DBFirewallID  string
	LBFirewallID  string
	CPFirewallID  string
	WPFirewallID  string
	SSHID         string // used to store the ssh id from CIVO
	Configuration *JsonStore
	SSH_Payload   *util.SSHPayload
}

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

type JsonStore struct {
	ClusterName string     `json:"clustername"`
	Region      string     `json:"region"`
	DBEndpoint  string     `json:"dbendpoint"`
	ServerToken string     `json:"servertoken"`
	SSHID       string     `json:"ssh_id"`
	InstanceIDs InstanceID `json:"instanceids"`
	NetworkIDs  NetworkID  `json:"networkids"`
}

func GetConfig(clusterName, region string) (configStore JsonStore, err error) {

	fileBytes, err := os.ReadFile(util.GetPath(util.CLUSTER_PATH, "civo", "ha", clusterName+" "+region, "info.json"))

	if err != nil {
		return
	}

	err = json.Unmarshal(fileBytes, &configStore)

	if err != nil {
		return
	}

	return
}

func saveConfig(logging log.Logger, clusterFolder string, configStore JsonStore) error {

	storeBytes, err := json.Marshal(configStore)
	if err != nil {
		return err
	}

	err = os.MkdirAll(util.GetPath(util.CLUSTER_PATH, "civo", "ha", clusterFolder), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.WriteFile(util.GetPath(util.CLUSTER_PATH, "civo", "ha", clusterFolder, "info.json"), storeBytes, 0640)
	if err != nil {
		return err
	}
	logging.Info("💾 configuration", "")
	return nil
}

type ConfigurationHandlers interface {
	ConfigWriterDBEndpoint(log.Logger, string) error
	ConfigWriterInstanceDatabase(log.Logger, string) error
	ConfigWriterServerToken(log.Logger, string) error
	ConfigWriterInstanceLoadBalancer(log.Logger, string) error
	ConfigWriterInstanceControlPlaneNodes(log.Logger, string) error
	ConfigWriterInstanceWorkerNodes(log.Logger, string) error
	ConfigWriterFirewallLoadBalancerNodes(log.Logger, string) error
	ConfigWriterFirewallControlPlaneNodes(log.Logger, string) error
	ConfigWriterFirewallWorkerNodes(log.Logger, string) error
	ConfigWriterFirewallDatabaseNodes(log.Logger, string) error
	ConfigWriterNetworkID(log.Logger, string) error
	ConfigWriterSSHID(log.Logger, string) error
}

// ConfigWriterDBEndpoint write Database endpoint to state management file
func (config *JsonStore) ConfigWriterDBEndpoint(logging log.Logger, endpoint string) error {
	config.DBEndpoint = endpoint
	return saveConfig(logging, config.ClusterName+" "+config.Region, *config)
}

// ConfigWriterSSHID write SSH keypairId which is uploaded to Civo to state management file
func (config *JsonStore) ConfigWriterSSHID(logging log.Logger, keypair_id string) error {
	config.SSHID = keypair_id
	return saveConfig(logging, config.ClusterName+" "+config.Region, *config)
}

// ConfigWriterNetworkID write NetworkID of created network Civo to state management file
func (config *JsonStore) ConfigWriterNetworkID(logging log.Logger, netID string) error {
	config.NetworkIDs.NetworkID = netID
	return saveConfig(logging, config.ClusterName+" "+config.Region, *config)
}

// ConfigWriterFirewallControlPlaneNodes write firewall_id of all controlplane's firewall to state management file
func (config *JsonStore) ConfigWriterFirewallControlPlaneNodes(logging log.Logger, fwID string) error {
	config.NetworkIDs.FirewallIDControlPlaneNode = fwID
	return saveConfig(logging, config.ClusterName+" "+config.Region, *config)
}

// ConfigWriterFirewallWorkerNodes write firewall_id of all workernode's firewall to state management file
func (config *JsonStore) ConfigWriterFirewallWorkerNodes(logging log.Logger, fwID string) error {
	config.NetworkIDs.FirewallIDWorkerNode = fwID
	return saveConfig(logging, config.ClusterName+" "+config.Region, *config)
}

// ConfigWriterFirewallLoadBalancerNodes write firewall_id of loadbalancer firewall to state management file
// TODO: Add more fine grained firewall rules
func (config *JsonStore) ConfigWriterFirewallLoadBalancerNodes(logging log.Logger, fwID string) error {
	config.NetworkIDs.FirewallIDLoadBalancerNode = fwID
	return saveConfig(logging, config.ClusterName+" "+config.Region, *config)
}

// ConfigWriterFirewallDatabaseNodes write firewall_id of database firewall to state management file
// TODO: Add more restrictive firewall rules
func (config *JsonStore) ConfigWriterFirewallDatabaseNodes(logging log.Logger, fwID string) error {
	config.NetworkIDs.FirewallIDDatabaseNode = fwID
	return saveConfig(logging, config.ClusterName+" "+config.Region, *config)
}

// ConfigWriterServerToken write the K3S_TOKEN to the state management file
func (config *JsonStore) ConfigWriterServerToken(logging log.Logger, token string) error {
	config.ServerToken = token
	return saveConfig(logging, config.ClusterName+" "+config.Region, *config)
}

// ConfigWriterInstanceDatabase write the instance_id of database VM to state management file
func (config *JsonStore) ConfigWriterInstanceDatabase(logging log.Logger, instanceID string) error {
	config.InstanceIDs.DatabaseNode = append(config.InstanceIDs.DatabaseNode, instanceID)
	return saveConfig(logging, config.ClusterName+" "+config.Region, *config)
}

// ConfigWriterInstanceLoadBalancer write the instance_id of loadbalancer VM to state management file
func (config *JsonStore) ConfigWriterInstanceLoadBalancer(logging log.Logger, instanceID string) error {
	config.InstanceIDs.LoadBalancerNode = append(config.InstanceIDs.LoadBalancerNode, instanceID)
	return saveConfig(logging, config.ClusterName+" "+config.Region, *config)
}

// ConfigWriterInstanceControlPlaneNodes write the instance_id of controlplane VMs to state management file
func (config *JsonStore) ConfigWriterInstanceControlPlaneNodes(logging log.Logger, instanceID string) error {
	config.InstanceIDs.ControlNodes = append(config.InstanceIDs.ControlNodes, instanceID)
	return saveConfig(logging, config.ClusterName+" "+config.Region, *config)
}

// ConfigWriterInstanceWorkerNodes write the instance_id of workernode VMs to state management file
func (config *JsonStore) ConfigWriterInstanceWorkerNodes(logging log.Logger, instanceID string) error {
	config.InstanceIDs.WorkerNodes = append(config.InstanceIDs.WorkerNodes, instanceID)
	return saveConfig(logging, config.ClusterName+" "+config.Region, *config)
}

// DeleteInstances deletes all the VMs
// deletes controlplane VMs, workerplane VMs, Database VM, Loadbalancer VM
func (obj *HAType) DeleteInstances(logging log.Logger) error {
	instances, err := ExtractInstances(obj.ClusterName, obj.Client.Region)
	if err != nil {
		return err
	}

	var errV error
	// if some controlplanes are down then there will be errors i.e if controlplanes are deleted
	for index, instanceID := range instances.ControlNodes {
		if err := obj.DeleteInstance(instanceID); err != nil {
			errV = err
			logging.Err(fmt.Sprintf("❌ [%d/%d] deleted controlplane instances", index+1, len(instances.ControlNodes)))
			continue
		}
		logging.Info(fmt.Sprintf("✅ [%d/%d] deleted controlplane instances", index+1, len(instances.ControlNodes)), instanceID)
	}
	if errV != nil {
		return errV
	}

	errV = nil
	for index, instanceID := range instances.WorkerNodes {
		if err := obj.DeleteInstance(instanceID); err != nil {
			errV = err
			logging.Err(fmt.Sprintf("❌ [%d/%d] deleted workerplane instances", index+1, len(instances.WorkerNodes)))
			continue
		}
		logging.Info(fmt.Sprintf("✅ [%d/%d] deleted workerplane instances", index+1, len(instances.WorkerNodes)), instanceID)
	}
	if errV != nil {
		return errV
	}

	errV = nil
	for index, instanceID := range instances.LoadBalancerNode {
		if err := obj.DeleteInstance(instanceID); err != nil {
			errV = err
			logging.Err(fmt.Sprintf("❌ [%d/%d] deleted loadbalancer instances", index+1, len(instances.LoadBalancerNode)))
			continue
		}
		logging.Info(fmt.Sprintf("✅ [%d/%d] deleted loadbalancer instances", index+1, len(instances.LoadBalancerNode)), instanceID)
	}
	if errV != nil {
		return errV
	}

	errV = nil
	for index, instanceID := range instances.DatabaseNode {
		if err := obj.DeleteInstance(instanceID); err != nil {
			errV = err
			logging.Err(fmt.Sprintf("❌ [%d/%d] deleted database instances", index+1, len(instances.DatabaseNode)))
			continue
		}
		logging.Info(fmt.Sprintf("✅ [%d/%d] deleted database instances", index+1, len(instances.DatabaseNode)), instanceID)
	}
	return errV
}

// DeleteNetworks deletes all network related objects
// deletes all firewalls, and the network
func (obj *HAType) DeleteNetworks(logging log.Logger) error {
	networks, err := ExtractNetworks(obj.ClusterName, obj.Client.Region)
	if err != nil {
		return err
	}

	if len(networks.FirewallIDControlPlaneNode) != 0 {
		err = obj.DeleteFirewall(networks.FirewallIDControlPlaneNode)
		if err != nil {
			logging.Err(fmt.Sprintf("❌ deleted controlplane firewall"))
		}
		logging.Info(fmt.Sprintf("✅ deleted controlplane firewall"), networks.FirewallIDControlPlaneNode)
	}

	time.Sleep(5 * time.Second)

	if len(networks.FirewallIDWorkerNode) != 0 {
		err = obj.DeleteFirewall(networks.FirewallIDWorkerNode)
		if err != nil {
			logging.Err(fmt.Sprintf("❌ deleted workerplane firewall"))
		}
		logging.Info(fmt.Sprintf("✅ deleted workerplane firewall"), networks.FirewallIDWorkerNode)
	}

	time.Sleep(5 * time.Second)

	if len(networks.FirewallIDDatabaseNode) != 0 {
		err = obj.DeleteFirewall(networks.FirewallIDDatabaseNode)
		if err != nil {
			logging.Err(fmt.Sprintf("❌ deleted database firewall"))
		}
		logging.Info(fmt.Sprintf("✅ deleted database firewall"), networks.FirewallIDDatabaseNode)
	}

	time.Sleep(5 * time.Second)

	if len(networks.FirewallIDLoadBalancerNode) != 0 {
		err = obj.DeleteFirewall(networks.FirewallIDLoadBalancerNode)
		if err != nil {
			logging.Err(fmt.Sprintf("❌ deleted loadbalancer firewall"))
		}
		logging.Info(fmt.Sprintf("✅ deleted loadbalancer firewall"), networks.FirewallIDLoadBalancerNode)
	}

	err = nil
	retry := 0
	retryTimeout := 2
	for retry < MAX_RETRY_COUNT {
		err = obj.DeleteNetwork(networks.NetworkID)
		if err == nil {
			break
		}
		retry++
		time.Sleep(time.Duration(retryTimeout) * time.Second)
		retryTimeout *= 2
		logging.Warn(fmt.Sprintln("❗ RETRYING ", err))
	}

	if retry == MAX_RETRY_COUNT {
		return fmt.Errorf("❌ deleted network")
	}

	logging.Info("✅ deleted network", "")

	return nil
}

// DeleteInstance delete a VM with instance_id
func (obj *HAType) DeleteInstance(instanceID string) error {
	_, err := obj.Client.DeleteInstance(instanceID)
	return err
}

// DeleteFirewall delete a firewall with firewall_id
func (obj *HAType) DeleteFirewall(firewallID string) error {
	_, err := obj.Client.DeleteFirewall(firewallID)
	return err
}

// DeleteNetwork delete the network with network_id
func (obj *HAType) DeleteNetwork(networkID string) error {
	_, err := obj.Client.DeleteNetwork(networkID)
	return err
}

// DeleteSSHKeyPair delete the SSH Keypair in CIVO
func (obj *HAType) DeleteSSHKeyPair() error {
	_, err := obj.Client.DeleteSSHKey(obj.SSHID)
	return err
}

// GetNetwork get network object with the provided name
func (obj *HAType) GetNetwork(networkName string) (net *civogo.Network, err error) {
	net, err = obj.Client.GetNetwork(networkName)
	return
}

// GetInstance get instance object with provided instance_id
func (obj *HAType) GetInstance(instanceID string) (inst *civogo.Instance, err error) {
	inst, err = obj.Client.GetInstance(instanceID)
	return
}

// CreateInstance create a instance with provided Configuration
// NOTE: initializationScript: if "" -> no default VM script when it is ready to serve
// else -> provide the script to run when the VM is ready (no need to SSH into to exec script)
// mention the `#!/bin/bash` for scripts
func (obj *HAType) CreateInstance(instanceName, firewallID, NodeSize, initializationScript string, public bool) (inst *civogo.Instance, err error) {
	publicIP := "create"
	if !public {
		publicIP = "none"
	}

	instanceConfig := &civogo.InstanceConfig{
		Hostname:         instanceName,
		InitialUser:      "root",
		Region:           obj.Client.Region,
		FirewallID:       firewallID,
		Size:             NodeSize,
		TemplateID:       obj.DiskImgID,
		NetworkID:        obj.NetworkID,
		Script:           initializationScript,
		SSHKeyID:         obj.SSHID,
		PublicIPRequired: publicIP,
	}

	inst, err = obj.Client.CreateInstance(instanceConfig)

	return
}

// CreateFirewall creates firewall with provided name and returns the firewall object
func (obj *HAType) CreateFirewall(firewallName string) (firew *civogo.FirewallResult, err error) {
	firewallConfig := &civogo.FirewallConfig{
		Name:      firewallName,
		Region:    obj.Client.Region,
		NetworkID: obj.NetworkID,
	}

	firew, err = obj.Client.NewFirewall(firewallConfig)

	return
}

// CreateNetwork creates network with provided name
func (obj *HAType) CreateNetwork(logging log.Logger, networkName string) error {
	net, err := obj.Client.NewNetwork(networkName)
	if err != nil {
		return err
	}
	obj.NetworkID = net.ID
	return obj.Configuration.ConfigWriterNetworkID(logging, net.ID)
}

// CreateSSHKeyPair upload's ssh keypair to CIVO using the Public Key generated by ssh-keygen
func (obj *HAType) CreateSSHKeyPair(logging log.Logger, publicKey string) error {
	sshRes, err := obj.Client.NewSSHKey(obj.ClusterName+"-"+strings.ToLower(obj.Client.Region)+"-ksctl-ha", publicKey)
	if err != nil {
		return err
	}
	obj.SSHID = sshRes.ID
	err = obj.Configuration.ConfigWriterSSHID(logging, sshRes.ID)
	return err
}

// SaveKubeconfig stores the kubeconfig to state management file
func (obj *HAType) SaveKubeconfig(logging log.Logger, kubeconfig string) error {
	folderName := obj.ClusterName + " " + obj.Client.Region
	err := os.MkdirAll(util.GetPath(util.CLUSTER_PATH, "civo", "ha", folderName), 0644)
	if err != nil && !os.IsExist(err) {
		return err
	}

	_, err = os.Create(util.GetPath(util.CLUSTER_PATH, "civo", "ha", folderName, "config"))
	if err != nil && !os.IsExist(err) {
		return err
	}

	file, err := os.OpenFile(util.GetPath(util.CLUSTER_PATH, "civo", "ha", folderName, "config"), os.O_WRONLY, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(kubeconfig))
	if err != nil {
		return err
	}
	logging.Info("💾 Kubeconfig", "")
	return nil
}

// ExtractInstances fetch all VMs instance_id from state management file
func ExtractInstances(clusterName, region string) (instIDs InstanceID, err error) {
	data, err := GetConfig(clusterName, region)
	if err != nil {
		err = fmt.Errorf("🚩 NO matching instance(s) found")
		return
	}

	instIDs = data.InstanceIDs
	return
}

// ExtractNetworks fetch the network_id from state management file
func ExtractNetworks(clusterName, region string) (instIDs NetworkID, err error) {
	data, err := GetConfig(clusterName, region)
	if err != nil {
		err = fmt.Errorf("🚩 NO matching network / firewall(s) found")
		return
	}

	instIDs = data.NetworkIDs
	return
}

// DeleteAllPaths
// WARNING: it is a destructive method
// removes all the info related to cluster (i.e. stat management file, configs and related info)
func DeleteAllPaths(clusterName, region string) error {
	return os.RemoveAll(util.GetPath(util.CLUSTER_PATH, "civo", "ha", clusterName+" "+region))
}

// UploadSSHKey it creates a ssh keypair saves to state management file and uploads it to CIVO
func (ha *HAType) UploadSSHKey(logging log.Logger) (err error) {
	path := util.GetPath(util.OTHER_PATH, "civo", "ha", ha.ClusterName+" "+ha.Client.Region)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return
	}
	keyPairToUpload, err := util.CreateSSHKeyPair("civo", ha.ClusterName+" "+ha.Client.Region)
	if err != nil {
		return
	}

	err = ha.CreateSSHKeyPair(logging, keyPairToUpload)

	// ------- Setting the ssh configs only the public ips used will change
	ha.SSH_Payload.UserName = "root"
	ha.SSH_Payload.PathPrivateKey = util.GetPath(util.SSH_PATH, "civo", "ha", ha.ClusterName+" "+ha.Client.Region)
	ha.SSH_Payload.Output = ""
	ha.SSH_Payload.PublicIP = ""
	// ------

	return
}
