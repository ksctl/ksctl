package ha_civo

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
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
			log.Printf("[ RETRYING ] %v\n", err)
		}
		time.Sleep(10 * time.Second) // waiting for ssh to get started
		currRetryCounter++
	}
	if currRetryCounter == MAX_RETRY_COUNT {
		return fmt.Errorf("[FATAL] COULDN'T RETRY: %v", err)
	}

	log.Println("CONFIGURING...")
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
			log.Printf("[ RETRYING ] %v\n", err)
		}
		time.Sleep(10 * time.Second) // waiting for ssh to get started
		currRetryCounter++
	}
	if currRetryCounter == MAX_RETRY_COUNT {
		return "", fmt.Errorf("[FATAL] COULDN'T RETRY: %v", err)
	}

	log.Println("CONFIGURING...")
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
	if len(instances) == 0 {
		return nil
	}
	for _, instanceID := range instances {
		if err := obj.DeleteInstance(instanceID); err != nil {
			return err
		}
	}
	return nil
}

func (obj *HAType) DeleteFirewalls() error {
	firewalls, err := ExtractFirewalls(obj.ClusterName, obj.Client.Region)
	if err != nil {
		return err
	}
	if len(firewalls) == 0 {
		return nil
	}
	for _, firewallID := range firewalls {
		if err := obj.DeleteFirewall(firewallID); err != nil {
			return err
		}
	}
	return nil
}

func (obj *HAType) DeleteNetworks() error {
	networks, err := ExtractNetwork(obj.ClusterName, obj.Client.Region)
	if err != nil {
		return err
	}
	if len(networks) == 0 {
		return nil
	}
	for _, networkID := range networks {
		if err := obj.DeleteNetwork(networkID); err != nil {
			return err
		}
	}
	return nil
}

func (obj *HAType) DeleteInstance(instanceID string) error {
	resp, err := obj.Client.DeleteInstance(instanceID)
	defer log.Println(resp)
	return err
}

func (obj *HAType) DeleteFirewall(firewallID string) error {
	resp, err := obj.Client.DeleteFirewall(firewallID)
	defer log.Println(resp)
	return err
}

func (obj *HAType) DeleteNetwork(networkID string) error {
	resp, err := obj.Client.DeleteNetwork(networkID)
	if errors.Is(civogo.DatabaseNetworkDeleteWithInstanceError, err) {
		time.Sleep(10 * time.Second)
		return obj.DeleteNetwork(networkID)
	} else {
		log.Println(resp)
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

func (obj *HAType) CreateInstance(instanceName, firewallID, NodeSize, initializationScript string) (inst *civogo.Instance, err error) {
	instanceConfig := &civogo.InstanceConfig{
		Hostname:    instanceName,
		InitialUser: "root",
		Region:      obj.Client.Region,
		FirewallID:  firewallID,
		Size:        NodeSize,
		TemplateID:  obj.DiskImgID,
		NetworkID:   obj.NetworkID,
		Script:      initializationScript,
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
	return obj.ConfigWriterNetwork(net)
}

func (obj *HAType) ConfigWriterInstance(instanceObj *civogo.Instance) error {
	// NOTE: location -> '~/.ksctl/config/ha-civo/name region/info/instances' file will contain all the instanceID
	//  location -> '~/.ksctl/config/ha-civo/name region/config' KUBECONFIG

	folderName := obj.ClusterName + " " + obj.Client.Region
	err := os.Mkdir(payload.GetPathCIVO(1, "ha-civo", folderName), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(payload.GetPathCIVO(1, "ha-civo", folderName, "info"), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	file, err := os.OpenFile(payload.GetPathCIVO(1, "ha-civo", folderName, "info", "instances"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(instanceObj.ID + " ")
	if err != nil {
		return err
	}

	return nil
}

func (obj *HAType) ConfigWriterFirewall(firewallObj *civogo.FirewallResult) error {
	// NOTE: location -> '~/.ksctl/config/ha-civo/name region/info/firewalls' file will contain all the instanceID
	//  location -> '~/.ksctl/config/ha-civo/name region/config' KUBECONFIG
	folderName := obj.ClusterName + " " + obj.Client.Region
	err := os.Mkdir(payload.GetPathCIVO(1, "ha-civo", folderName), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(payload.GetPathCIVO(1, "ha-civo", folderName, "info"), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	file, err := os.OpenFile(payload.GetPathCIVO(1, "ha-civo", folderName, "info", "firewalls"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(firewallObj.ID + " ")
	if err != nil {
		return err
	}

	return nil
}

func (obj *HAType) ConfigWriterNetwork(networkObj *civogo.NetworkResult) error {
	// NOTE: location -> '~/.ksctl/config/ha-civo/name region/info/network' file will contain all the instanceID
	//  location -> '~/.ksctl/config/ha-civo/name region/config' KUBECONFIG

	folderName := obj.ClusterName + " " + obj.Client.Region
	err := os.Mkdir(payload.GetPathCIVO(1, "ha-civo", folderName), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(payload.GetPathCIVO(1, "ha-civo", folderName, "info"), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	file, err := os.OpenFile(payload.GetPathCIVO(1, "ha-civo", folderName, "info", "network"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(networkObj.ID + " ")
	if err != nil {
		return err
	}

	return nil
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
	return nil
}

func ExtractInstances(clusterName, region string) ([]string, error) {
	data, err := os.ReadFile(payload.GetPathCIVO(1, "ha-civo", clusterName+" "+region, "info", "instances"))
	if err != nil {
		return nil, fmt.Errorf("NO matching instance(s) found")
	}

	arr := strings.Split(strings.TrimSpace(string(data)), " ")

	return arr, nil
}

func ExtractFirewalls(clusterName, region string) ([]string, error) {
	data, err := os.ReadFile(payload.GetPathCIVO(1, "ha-civo", clusterName+" "+region, "info", "firewalls"))
	if err != nil {
		return nil, fmt.Errorf("NO matching firewall(s) found")
	}

	arr := strings.Split(strings.TrimSpace(string(data)), " ")

	return arr, nil
}

func ExtractNetwork(clusterName, region string) ([]string, error) {
	data, err := os.ReadFile(payload.GetPathCIVO(1, "ha-civo", clusterName+" "+region, "info", "network"))
	if err != nil {
		return nil, fmt.Errorf("NO matching network(s) found")
	}

	arr := strings.Split(strings.TrimSpace(string(data)), " ")

	return arr, nil
}

func DeleteAllPaths(clusterName, region string) error {
	return os.RemoveAll(payload.GetPathCIVO(1, "ha-civo", clusterName+" "+region))
}

type HACollection interface {
	DeleteInstances() error
	DeleteFirewalls() error
	DeleteNetworks() error

	DeleteInstance(string) error
	DeleteFirewall(string) error
	DeleteNetwork(string) error

	GetNetwork(string) (*civogo.Network, error)
	GetInstance(string) (*civogo.Instance, error)

	CreateFirewall(string) (*civogo.FirewallResult, error)
	CreateNetwork(string) error
	CreateInstance(string, string, string, string) (*civogo.Instance, error)

	ConfigWriterInstance(*civogo.Instance) error
	ConfigWriterFirewall(*civogo.FirewallResult) error
	ConfigWriterNetwork(*civogo.NetworkResult) error
	SaveKubeconfig(string) error

	CreateLoadbalancer() (*civogo.Instance, error)
	CreateControlPlane(int) (*civogo.Instance, error)
	CreateWorkerNode(int, string, string) (*civogo.Instance, error)
	CreateDatabase() (string, error)
}

type HAType struct {
	Client       *civogo.Client // CIVO client obj
	ClusterName  string         // clusterName provided by the user
	DiskImgID    string         // disk Img ID for ubuntu
	NetworkID    string         // network id
	NodeSize     string         // e.x. g3.medium
	DBFirewallID string
	LBFirewallID string
	CPFirewallID string
	WPFirewallID string
}
