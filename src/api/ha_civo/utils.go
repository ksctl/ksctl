package ha_civo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/src/api/payload"
	"golang.org/x/crypto/ssh"
)

const (
	SSH_PAUSE_IN_SECONDS = 20
	MAX_RETRY_COUNT      = 6
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
	Configuration *JsonStore
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
	InstanceIDs InstanceID `json:"instanceids"`
	NetworkIDs  NetworkID  `json:"networkids"`
}

func GetConfig(clusterName, region string) (configStore JsonStore, err error) {

	fileBytes, err := ioutil.ReadFile(payload.GetPathCIVO(1, "ha-civo", clusterName+" "+region, "info.json"))

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

	err = os.Mkdir(payload.GetPathCIVO(1, "ha-civo", clusterFolder), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// _, err = os.Create(payload.GetPathCIVO(1, "ha-civo", clusterFolder, "config"))
	// if err != nil && !os.IsExist(err) {
	// 	return err
	// }

	err = ioutil.WriteFile(payload.GetPathCIVO(1, "ha-civo", clusterFolder, "info.json"), storeBytes, 0640)
	if err != nil {
		return err
	}
	log.Println("💾 📃 configuration")
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
}

func (config *JsonStore) ConfigWriterDBEndpoint(endpoint string) error {
	config.DBEndpoint = endpoint
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

func ExecWithoutOutput(publicIP, password, script string, fastMode bool) error {

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		//HostKeyCallback: hostKeyCallback,
		// FIXME: Insecure Ignore should be replaced with secure
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if !fastMode {
		time.Sleep(SSH_PAUSE_IN_SECONDS * time.Second)
	}

	var err error
	var conn *ssh.Client
	currRetryCounter := 0

	for currRetryCounter < MAX_RETRY_COUNT {
		conn, err = ssh.Dial("tcp", publicIP+":22", config)
		if err == nil {
			break
		} else {
			log.Printf("❗ RETRYING %v\n", err)
		}
		time.Sleep(10 * time.Second) // waiting for ssh to get started
		currRetryCounter++
	}
	if currRetryCounter == MAX_RETRY_COUNT {
		return fmt.Errorf("🚨 💀 COULDN'T RETRY: %v", err)
	}

	log.Println("🤖 📃 Exec Scripts")
	defer conn.Close()

	session, err := conn.NewSession()

	if err != nil {
		return err
	}

	defer session.Close()

	if err := session.Run(script); err != nil {
		return err
	}

	return nil
}

func ExecWithOutput(publicIP, password, script string, fastMode bool) (string, error) {

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		//HostKeyCallback: hostKeyCallback,
		// FIXME: Insecure Ignore should be replaced with secure
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if !fastMode {
		time.Sleep(SSH_PAUSE_IN_SECONDS * time.Second)
	}
	var err error
	var conn *ssh.Client
	currRetryCounter := 0

	for currRetryCounter < MAX_RETRY_COUNT {
		conn, err = ssh.Dial("tcp", publicIP+":22", config)
		if err == nil {
			break
		} else {
			log.Printf("❗ RETRYING %v\n", err)
		}
		time.Sleep(10 * time.Second) // waiting for ssh to get started
		currRetryCounter++
	}
	if currRetryCounter == MAX_RETRY_COUNT {
		return "", fmt.Errorf("🚨 💀 COULDN'T RETRY: %v", err)
	}

	log.Println("🤖 📃 Exec Scripts")
	defer conn.Close()

	session, err := conn.NewSession()

	if err != nil {
		return "", err
	}

	defer session.Close()

	var buff bytes.Buffer
	session.Stdout = &buff

	if err := session.Run(script); err != nil {
		return "", err
	}

	return buff.String(), nil
}

func (obj *HAType) DeleteInstances() error {
	instances, err := ExtractInstances(obj.ClusterName, obj.Client.Region)
	if err != nil {
		return err
	}

	for index, instanceID := range instances.ControlNodes {
		if err := obj.DeleteInstance(instanceID); err != nil {
			log.Println(fmt.Sprintf("❌ [%d/%d] deleted controlplane instances", index+1, len(instances.ControlNodes)))
			return err
		}
		log.Println(fmt.Sprintf("✅ [%d/%d] deleted controlplane instances", index+1, len(instances.ControlNodes)))
	}

	for index, instanceID := range instances.WorkerNodes {
		if err := obj.DeleteInstance(instanceID); err != nil {
			log.Println(fmt.Sprintf("❌ [%d/%d] deleted workerplane instances", index+1, len(instances.WorkerNodes)))
			return err
		}
		log.Println(fmt.Sprintf("✅ [%d/%d] deleted workerplane instances", index+1, len(instances.WorkerNodes)))
	}

	for index, instanceID := range instances.LoadBalancerNode {
		if err := obj.DeleteInstance(instanceID); err != nil {
			log.Println(fmt.Sprintf("❌ [%d/%d] deleted loadbalancer instances", index+1, len(instances.LoadBalancerNode)))
			return err
		}
		log.Println(fmt.Sprintf("✅ [%d/%d] deleted loadbalancer instances", index+1, len(instances.LoadBalancerNode)))
	}

	for index, instanceID := range instances.DatabaseNode {
		if err := obj.DeleteInstance(instanceID); err != nil {
			log.Println(fmt.Sprintf("❌ [%d/%d] deleted database instances", index+1, len(instances.DatabaseNode)))
			return err
		}
		log.Println(fmt.Sprintf("✅ [%d/%d] deleted database instances", index+1, len(instances.DatabaseNode)))
	}
	return nil
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

	time.Sleep(5 * time.Second)

	err = obj.DeleteNetwork(networks.NetworkID)
	if err != nil {
		log.Println(fmt.Sprintf("❌ deleted network"))
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
	if errors.Is(civogo.DatabaseNetworkDeleteWithInstanceError, err) {
		time.Sleep(10 * time.Second)
		return obj.DeleteNetwork(networkID)
	}
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

func (obj *HAType) SaveKubeconfig(kubeconfig string) error {
	folderName := obj.ClusterName + " " + obj.Client.Region
	err := os.Mkdir(payload.GetPathCIVO(1, "ha-civo", folderName), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	_, err = os.Create(payload.GetPathCIVO(1, "ha-civo", folderName, "config"))
	if err != nil && !os.IsExist(err) {
		return err
	}

	file, err := os.OpenFile(payload.GetPathCIVO(1, "ha-civo", folderName, "config"), os.O_WRONLY, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(kubeconfig))
	if err != nil {
		return err
	}
	log.Println("✅ 📃 Kubeconfig")
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
	return os.RemoveAll(payload.GetPathCIVO(1, "ha-civo", clusterName+" "+region))
}
