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
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

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
	DeleteInstances() error
	DeleteNetworks() error

	DeleteInstance(string) error
	DeleteFirewall(string) error
	DeleteNetwork(string) error

	GetNetwork(string) (*civogo.Network, error)
	GetInstance(string) (*civogo.Instance, error)

	CreateFirewall(string) (*civogo.FirewallResult, error)
	CreateNetwork(string) error
	CreateInstance(string, string, string, string, bool) (*civogo.Instance, error)

	SaveKubeconfig(string) error

	CreateLoadbalancer() (*civogo.Instance, error)
	CreateControlPlane(int) (*civogo.Instance, error)
	CreateWorkerNode(int, string, string) (*civogo.Instance, error)
	CreateDatabase() (string, error)
	GetTokenFromCP_1(*civogo.Instance) string

	UploadSSHKey() error
	DeleteSSHKeyPair() error
	ConfigLoadBalancer(*civogo.Instance, []string) error
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

	fileBytes, err := ioutil.ReadFile(util.GetPath(1, "civo", "ha", clusterName+" "+region, "info.json"))

	if err != nil {
		return
	}

	err = json.Unmarshal(fileBytes, &configStore)

	if err != nil {
		return
	}

	return
}

func saveConfig(clusterFolder string, configStore JsonStore) error {

	storeBytes, err := json.Marshal(configStore)
	if err != nil {
		return err
	}

	err = os.MkdirAll(util.GetPath(1, "civo", "ha", clusterFolder), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = ioutil.WriteFile(util.GetPath(1, "civo", "ha", clusterFolder, "info.json"), storeBytes, 0640)
	if err != nil {
		return err
	}
	log.Println("💾 configuration")
	return nil
}

type ConfigurationHandlers interface {
	ConfigWriterDBEndpoint(string) error
	ConfigWriterInstanceDatabase(string) error
	ConfigWriterServerToken(string) error
	ConfigWriterInstanceLoadBalancer(string) error
	ConfigWriterInstanceControlPlaneNodes(string) error
	ConfigWriterInstanceWorkerNodes(string) error
	ConfigWriterFirewallLoadBalancerNodes(string) error
	ConfigWriterFirewallControlPlaneNodes(string) error
	ConfigWriterFirewallWorkerNodes(string) error
	ConfigWriterFirewallDatabaseNodes(string) error
	ConfigWriterNetworkID(string) error
	ConfigWriterSSHID(string) error
}

func (config *JsonStore) ConfigWriterDBEndpoint(endpoint string) error {
	config.DBEndpoint = endpoint
	return saveConfig(config.ClusterName+" "+config.Region, *config)
}

func (config *JsonStore) ConfigWriterSSHID(keypair_id string) error {
	config.SSHID = keypair_id
	return saveConfig(config.ClusterName+" "+config.Region, *config)
}

func (config *JsonStore) ConfigWriterNetworkID(netID string) error {
	config.NetworkIDs.NetworkID = netID
	return saveConfig(config.ClusterName+" "+config.Region, *config)
}

func (config *JsonStore) ConfigWriterFirewallControlPlaneNodes(fwID string) error {
	config.NetworkIDs.FirewallIDControlPlaneNode = fwID
	return saveConfig(config.ClusterName+" "+config.Region, *config)
}

func (config *JsonStore) ConfigWriterFirewallWorkerNodes(fwID string) error {
	config.NetworkIDs.FirewallIDWorkerNode = fwID
	return saveConfig(config.ClusterName+" "+config.Region, *config)
}

func (config *JsonStore) ConfigWriterFirewallLoadBalancerNodes(fwID string) error {
	config.NetworkIDs.FirewallIDLoadBalancerNode = fwID
	return saveConfig(config.ClusterName+" "+config.Region, *config)
}

func (config *JsonStore) ConfigWriterFirewallDatabaseNodes(fwID string) error {
	config.NetworkIDs.FirewallIDDatabaseNode = fwID
	return saveConfig(config.ClusterName+" "+config.Region, *config)
}

func (config *JsonStore) ConfigWriterServerToken(token string) error {
	config.ServerToken = token
	return saveConfig(config.ClusterName+" "+config.Region, *config)
}

func (config *JsonStore) ConfigWriterInstanceDatabase(instanceID string) error {
	config.InstanceIDs.DatabaseNode = append(config.InstanceIDs.DatabaseNode, instanceID)
	return saveConfig(config.ClusterName+" "+config.Region, *config)
}

func (config *JsonStore) ConfigWriterInstanceLoadBalancer(instanceID string) error {
	config.InstanceIDs.LoadBalancerNode = append(config.InstanceIDs.LoadBalancerNode, instanceID)
	return saveConfig(config.ClusterName+" "+config.Region, *config)
}

func (config *JsonStore) ConfigWriterInstanceControlPlaneNodes(instanceID string) error {
	config.InstanceIDs.ControlNodes = append(config.InstanceIDs.ControlNodes, instanceID)
	return saveConfig(config.ClusterName+" "+config.Region, *config)
}

func (config *JsonStore) ConfigWriterInstanceWorkerNodes(instanceID string) error {
	config.InstanceIDs.WorkerNodes = append(config.InstanceIDs.WorkerNodes, instanceID)
	return saveConfig(config.ClusterName+" "+config.Region, *config)
}

func (obj *HAType) DeleteInstances() error {
	instances, err := ExtractInstances(obj.ClusterName, obj.Client.Region)
	if err != nil {
		return err
	}

	var errV error
	// if some controlplanes are down then there will be errors i.e if controlplanes are deleted
	for index, instanceID := range instances.ControlNodes {
		if err := obj.DeleteInstance(instanceID); err != nil {
			errV = err
			log.Println(fmt.Sprintf("❌ [%d/%d] deleted controlplane instances", index+1, len(instances.ControlNodes)))
			continue
		}
		log.Println(fmt.Sprintf("✅ [%d/%d] deleted controlplane instances", index+1, len(instances.ControlNodes)))
	}
	if errV != nil {
		return errV
	}

	errV = nil
	for index, instanceID := range instances.WorkerNodes {
		if err := obj.DeleteInstance(instanceID); err != nil {
			errV = err
			log.Println(fmt.Sprintf("❌ [%d/%d] deleted workerplane instances", index+1, len(instances.WorkerNodes)))
			continue
		}
		log.Println(fmt.Sprintf("✅ [%d/%d] deleted workerplane instances", index+1, len(instances.WorkerNodes)))
	}
	if errV != nil {
		return errV
	}

	errV = nil
	for index, instanceID := range instances.LoadBalancerNode {
		if err := obj.DeleteInstance(instanceID); err != nil {
			errV = err
			log.Println(fmt.Sprintf("❌ [%d/%d] deleted loadbalancer instances", index+1, len(instances.LoadBalancerNode)))
			continue
		}
		log.Println(fmt.Sprintf("✅ [%d/%d] deleted loadbalancer instances", index+1, len(instances.LoadBalancerNode)))
	}
	if errV != nil {
		return errV
	}

	errV = nil
	for index, instanceID := range instances.DatabaseNode {
		if err := obj.DeleteInstance(instanceID); err != nil {
			errV = err
			log.Println(fmt.Sprintf("❌ [%d/%d] deleted database instances", index+1, len(instances.DatabaseNode)))
			continue
		}
		log.Println(fmt.Sprintf("✅ [%d/%d] deleted database instances", index+1, len(instances.DatabaseNode)))
	}
	return errV
}

func (obj *HAType) DeleteNetworks() error {
	networks, err := ExtractNetworks(obj.ClusterName, obj.Client.Region)
	if err != nil {
		return err
	}

	if len(networks.FirewallIDControlPlaneNode) != 0 {
		err = obj.DeleteFirewall(networks.FirewallIDControlPlaneNode)
		if err != nil {
			log.Println(fmt.Sprintf("❌ deleted controlplane firewall"))
		}
		log.Println(fmt.Sprintf("✅ deleted controlplane firewall"))
	}

	time.Sleep(5 * time.Second)

	if len(networks.FirewallIDWorkerNode) != 0 {
		err = obj.DeleteFirewall(networks.FirewallIDWorkerNode)
		if err != nil {
			log.Println(fmt.Sprintf("❌ deleted workerplane firewall"))
		}
		log.Println(fmt.Sprintf("✅ deleted workerplane firewall"))
	}

	time.Sleep(5 * time.Second)

	if len(networks.FirewallIDDatabaseNode) != 0 {
		err = obj.DeleteFirewall(networks.FirewallIDDatabaseNode)
		if err != nil {
			log.Println(fmt.Sprintf("❌ deleted database firewall"))
		}
		log.Println(fmt.Sprintf("✅ deleted database firewall"))
	}

	time.Sleep(5 * time.Second)

	if len(networks.FirewallIDLoadBalancerNode) != 0 {
		err = obj.DeleteFirewall(networks.FirewallIDLoadBalancerNode)
		if err != nil {
			log.Println(fmt.Sprintf("❌ deleted loadbalancer firewall"))
		}
		log.Println(fmt.Sprintf("✅ deleted loadbalancer firewall"))
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
		log.Println("❗ RETRYING ", err)
	}

	if retry == MAX_RETRY_COUNT {
		return fmt.Errorf("❌ deleted network")
	}

	log.Println(fmt.Sprintf("✅ deleted network"))

	return nil
}

func (obj *HAType) DeleteInstance(instanceID string) error {
	_, err := obj.Client.DeleteInstance(instanceID)
	return err
}

func (obj *HAType) DeleteFirewall(firewallID string) error {
	_, err := obj.Client.DeleteFirewall(firewallID)
	return err
}

func (obj *HAType) DeleteNetwork(networkID string) error {
	_, err := obj.Client.DeleteNetwork(networkID)
	return err
}

func (obj *HAType) DeleteSSHKeyPair() error {
	_, err := obj.Client.DeleteSSHKey(obj.SSHID)
	return err
}

func (obj *HAType) GetNetwork(networkName string) (net *civogo.Network, err error) {
	net, err = obj.Client.GetNetwork(networkName)
	return
}

func (obj *HAType) GetInstance(instanceID string) (inst *civogo.Instance, err error) {
	inst, err = obj.Client.GetInstance(instanceID)
	return
}

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

func (obj *HAType) CreateFirewall(firewallName string) (firew *civogo.FirewallResult, err error) {
	firewallConfig := &civogo.FirewallConfig{
		Name:      firewallName,
		Region:    obj.Client.Region,
		NetworkID: obj.NetworkID,
	}

	firew, err = obj.Client.NewFirewall(firewallConfig)

	return
}

func (obj *HAType) CreateNetwork(networkName string) error {
	net, err := obj.Client.NewNetwork(networkName)
	if err != nil {
		return err
	}
	obj.NetworkID = net.ID
	return obj.Configuration.ConfigWriterNetworkID(net.ID)
}

func (obj *HAType) CreateSSHKeyPair(publicKey string) error {
	sshRes, err := obj.Client.NewSSHKey(obj.ClusterName+"-"+strings.ToLower(obj.Client.Region)+"-ksctl-ha", publicKey)
	if err != nil {
		return err
	}
	obj.SSHID = sshRes.ID
	err = obj.Configuration.ConfigWriterSSHID(sshRes.ID)
	return err
}

func (obj *HAType) SaveKubeconfig(kubeconfig string) error {
	folderName := obj.ClusterName + " " + obj.Client.Region
	err := os.MkdirAll(util.GetPath(1, "civo", "ha", folderName), 0644)
	if err != nil && !os.IsExist(err) {
		return err
	}

	_, err = os.Create(util.GetPath(1, "civo", "ha", folderName, "config"))
	if err != nil && !os.IsExist(err) {
		return err
	}

	file, err := os.OpenFile(util.GetPath(1, "civo", "ha", folderName, "config"), os.O_WRONLY, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(kubeconfig))
	if err != nil {
		return err
	}
	log.Println("💾 Kubeconfig")
	return nil
}

func ExtractInstances(clusterName, region string) (instIDs InstanceID, err error) {
	data, err := GetConfig(clusterName, region)
	if err != nil {
		err = fmt.Errorf("🚩 NO matching instance(s) found")
		return
	}

	instIDs = data.InstanceIDs
	return
}

func ExtractNetworks(clusterName, region string) (instIDs NetworkID, err error) {
	data, err := GetConfig(clusterName, region)
	if err != nil {
		err = fmt.Errorf("🚩 NO matching network / firewall(s) found")
		return
	}

	instIDs = data.NetworkIDs
	return
}

func DeleteAllPaths(clusterName, region string) error {
	return os.RemoveAll(util.GetPath(1, "civo", "ha", clusterName+" "+region))
}

// UploadSSHKey it creates a ssh keypair saves it locally and uploads it to CIVO
func (ha *HAType) UploadSSHKey() (err error) {
	path := util.GetPath(util.OTHER_PATH, "civo", "ha", ha.ClusterName+" "+ha.Client.Region)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return
	}
	keyPairToUpload, err := util.CreateSSHKeyPair("civo", ha.ClusterName, ha.Client.Region)
	if err != nil {
		return
	}

	err = ha.CreateSSHKeyPair(keyPairToUpload)

	// ------- Setting the ssh configs only the public ips used will change
	ha.SSH_Payload.UserName = "root"
	ha.SSH_Payload.PathPrivateKey = util.GetSSHPath("civo", "ha", ha.ClusterName+" "+ha.Client.Region)
	ha.SSH_Payload.Output = ""
	ha.SSH_Payload.PublicIP = ""
	// ------

	return
}
