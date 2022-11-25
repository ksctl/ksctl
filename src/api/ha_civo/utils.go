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

func DeleteInstances(client *civogo.Client, clusterName string) error {
	instances, err := ExtractInstances(clusterName, client.Region)
	if err != nil {
		return err
	}
	if len(instances) == 0 {
		return nil
	}
	for _, instanceID := range instances {
		if err := DeleteInstance(client, instanceID); err != nil {
			return err
		}
	}
	return nil
}

func DeleteFirewalls(client *civogo.Client, clusterName string) error {
	firewalls, err := ExtractFirewalls(clusterName, client.Region)
	if err != nil {
		return err
	}
	if len(firewalls) == 0 {
		return nil
	}
	for _, firewallID := range firewalls {
		if err := DeleteFirewall(client, firewallID); err != nil {
			return err
		}
	}
	return nil
}

func DeleteNetworks(client *civogo.Client, clusterName string) error {
	networks, err := ExtractNetwork(clusterName, client.Region)
	if err != nil {
		return err
	}
	if len(networks) == 0 {
		return nil
	}
	for _, networkID := range networks {
		if err := DeleteNetwork(client, networkID); err != nil {
			return err
		}
	}
	return nil
}

func DeleteInstance(client *civogo.Client, instanceID string) error {
	resp, err := client.DeleteInstance(instanceID)
	defer log.Println(resp)
	return err
}

func DeleteFirewall(client *civogo.Client, firewallID string) error {
	resp, err := client.DeleteFirewall(firewallID)
	defer log.Println(resp)
	return err
}

func DeleteNetwork(client *civogo.Client, networkID string) error {
	resp, err := client.DeleteNetwork(networkID)
	if errors.Is(civogo.DatabaseNetworkDeleteWithInstanceError, err) {
		time.Sleep(10 * time.Second)
		return DeleteNetwork(client, networkID)
	} else {
		log.Println(resp)
	}
	return err
}

func GetNetwork(client *civogo.Client, networkName string) (net *civogo.Network, err error) {
	net, err = client.GetNetwork(networkName)
	return
}

func GetInstance(client *civogo.Client, instanceID string) (inst *civogo.Instance, err error) {
	inst, err = client.GetInstance(instanceID)
	return
}

func CreateInstance(client *civogo.Client, instanceName, firewallID, diskImgID, nodeSize, networkID string, initializationScript string) (inst *civogo.Instance, err error) {
	instanceConfig := &civogo.InstanceConfig{
		Hostname:    instanceName,
		InitialUser: "root",
		Region:      client.Region,
		FirewallID:  firewallID,
		Size:        nodeSize,
		TemplateID:  diskImgID,
		NetworkID:   networkID,
		Script:      initializationScript,
	}

	inst, err = client.CreateInstance(instanceConfig)

	return
}

func CreateFirewall(client *civogo.Client, firewallName, networkID string) (firew *civogo.FirewallResult, err error) {
	firewallConfig := &civogo.FirewallConfig{
		Name:      firewallName,
		Region:    client.Region,
		NetworkID: networkID,
	}

	firew, err = client.NewFirewall(firewallConfig)

	return
}

func CreateNetwork(client *civogo.Client, networkName string) (net *civogo.NetworkResult, err error) {
	net, err = client.NewNetwork(networkName)
	return
}

func ConfigWriterInstance(instanceObj *civogo.Instance, clusterName, region string) error {
	// NOTE: location -> '~/.ksctl/config/ha-civo/name region/info/instances' file will contain all the instanceID
	//  location -> '~/.ksctl/config/ha-civo/name region/config' KUBECONFIG

	folderName := clusterName + " " + region
	err := os.Mkdir(GetPath(1, folderName), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(GetPath(1, folderName, "info"), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	file, err := os.OpenFile(GetPath(1, folderName, "info", "instances"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0750)
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

func ConfigWriterFirewall(firewallObj *civogo.FirewallResult, clusterName, region string) error {
	// NOTE: location -> '~/.ksctl/config/ha-civo/name region/info/firewalls' file will contain all the instanceID
	//  location -> '~/.ksctl/config/ha-civo/name region/config' KUBECONFIG
	folderName := clusterName + " " + region
	err := os.Mkdir(GetPath(1, folderName), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(GetPath(1, folderName, "info"), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	file, err := os.OpenFile(GetPath(1, folderName, "info", "firewalls"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0750)
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

func ConfigWriterNetwork(networkObj *civogo.NetworkResult, clusterName, region string) error {
	// NOTE: location -> '~/.ksctl/config/ha-civo/name region/info/network' file will contain all the instanceID
	//  location -> '~/.ksctl/config/ha-civo/name region/config' KUBECONFIG

	folderName := clusterName + " " + region
	err := os.Mkdir(GetPath(1, folderName), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(GetPath(1, folderName, "info"), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	file, err := os.OpenFile(GetPath(1, folderName, "info", "network"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0750)
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

func SaveKubeconfig(clusterName, region, kubeconfig string) error {
	folderName := clusterName + " " + region
	err := os.Mkdir(GetPath(1, folderName), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	_, err = os.Create(GetPath(1, folderName, "config"))
	if err != nil && !os.IsExist(err) {
		return err
	}

	file, err := os.OpenFile(GetPath(1, folderName, "config"), os.O_WRONLY, 0750)
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
	data, err := os.ReadFile(GetPath(1, clusterName+" "+region, "info", "instances"))
	if err != nil {
		return nil, fmt.Errorf("NO matching instance(s) found")
	}

	arr := strings.Split(strings.TrimSpace(string(data)), " ")

	return arr, nil
}

func ExtractFirewalls(clusterName, region string) ([]string, error) {
	data, err := os.ReadFile(GetPath(1, clusterName+" "+region, "info", "firewalls"))
	if err != nil {
		return nil, fmt.Errorf("NO matching firewall(s) found")
	}

	arr := strings.Split(strings.TrimSpace(string(data)), " ")

	return arr, nil
}

func ExtractNetwork(clusterName, region string) ([]string, error) {
	data, err := os.ReadFile(GetPath(1, clusterName+" "+region, "info", "network"))
	if err != nil {
		return nil, fmt.Errorf("NO matching network(s) found")
	}

	arr := strings.Split(strings.TrimSpace(string(data)), " ")

	return arr, nil
}

func DeleteAllPaths(clusterName, region string) error {
	return os.RemoveAll(GetPath(1, clusterName+" "+region))
}
