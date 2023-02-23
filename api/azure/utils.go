package azure

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	util "github.com/kubesimplify/ksctl/api/utils"
	"golang.org/x/net/context"
)

type AzureOperations interface {
	CreateCluster() error
	DeleteCluster() error

	AddMoreWorkerNodes() error
	DeleteSomeWorkerNodes() error
}

type AzureStateVMs struct {
	Names                    []string `json:"names"`
	NetworkSecurityGroupName string   `json:"network_security_group_name"`
	NetworkSecurityGroupID   string   `json:"network_security_group_id"`
	DiskNames                []string `json:"disk_names"`
	PublicIPNames            []string `json:"public_ip_names"`
	PrivateIPs               []string `json:"private_ips"`
	PublicIPs                []string `json:"public_ips"`
	NetworkInterfaceNames    []string `json:"network_interface_names"`
}

type AzureStateVM struct {
	Name                     string `json:"name"`
	NetworkSecurityGroupName string `json:"network_security_group_name"`
	NetworkSecurityGroupID   string `json:"network_security_group_id"`
	DiskName                 string `json:"disk_name"`
	PublicIPName             string `json:"public_ip_name"`
	NetworkInterfaceName     string `json:"network_interface_name"`
	PrivateIP                string `json:"private_ip"`
	PublicIP                 string `json:"public_ip"`
}

type AzureStateCluster struct {
	ClusterName       string `json:"cluster_name"`
	ResourceGroupName string `json:"resource_group_name"`
	SSHKeyName        string `json:"ssh_key_name"`
	DBEndpoint        string `json:"database_endpoint"`
	K3sToken          string `json:"k3s_token"`

	SubnetName         string `json:"subnet_name"`
	SubnetID           string `json:"subnet_id"`
	VirtualNetworkName string `json:"virtual_network_name"`
	VirtualNetworkID   string `json:"virtual_network_id"`

	InfoControlPlanes AzureStateVMs `json:"info_control_planes"`
	InfoWorkerPlanes  AzureStateVMs `json:"info_worker_planes"`
	InfoDatabase      AzureStateVM  `json:"info_database"`
	InfoLoadBalancer  AzureStateVM  `json:"info_load_balancer"`
}

type AzureInfra interface {

	// azure resources
	CreateResourceGroup(context.Context) error
	DeleteResourceGroup(context.Context) error
	DeleteDisk(context.Context, string) error
	DeleteSubnet(context.Context) error
	CreateSubnet(context.Context, string) (*armnetwork.Subnet, error)
	CreateVM(context.Context, string, string, string) (*armcompute.VirtualMachine, error)
	DeleteVM(context.Context, string) error
	CreateVirtualNetwork(context.Context, string) (*armnetwork.VirtualNetwork, error)
	DeleteVirtualNetwork(context.Context) error
	CreateNSG(context.Context, string, []*armnetwork.SecurityRule) (*armnetwork.SecurityGroup, error)
	DeleteNSG(context.Context, string) error
	DeleteNetworkInterface(context.Context, string) error
	CreateNetworkInterface(context.Context, string, string, string, string, string) (*armnetwork.Interface, error)
	DeletePublicIP(context.Context, string) error
	CreatePublicIP(context.Context, string) (*armnetwork.PublicIPAddress, error)

	UploadSSHKey(context.Context) (err error)
	DeleteSSHKey(context.Context) error

	// state file managemenet
	ConfigReader() error
	ConfigWriter() error

	// kubeconfig file
	kubeconfigWriter(string) error
	kubeconfigReader() ([]byte, error)
}

func (config *AzureProvider) ConfigWriter(clusterType string) error {
	return util.SaveState(config.Config, "azure", clusterType, config.ClusterName+" "+config.Config.ResourceGroupName+" "+config.Region)
}

func isPresent(kind string, obj AzureProvider) bool {
	path := util.GetPath(util.CLUSTER_PATH, "azure", kind, obj.ClusterName+" "+obj.Config.ResourceGroupName+" "+obj.Region, "info.json")
	_, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func (config *AzureProvider) ConfigReader(clusterType string) error {
	data, err := util.GetState("azure", clusterType, config.ClusterName+" "+config.Config.ResourceGroupName+" "+config.Region)
	if err != nil {
		return err
	}
	// Convert the map to JSON
	jsonData, _ := json.Marshal(data)

	// Convert the JSON to a struct
	var structData AzureStateCluster
	json.Unmarshal(jsonData, &structData)

	config.Config = &structData
	config.ClusterName = config.Config.ClusterName
	return nil
}

func setRequiredENV_VAR(ctx context.Context, cred *AzureProvider) error {
	tokens, err := util.GetCred("azure")
	if err != nil {
		return err
	}
	cred.SubscriptionID = tokens["subscription_id"]
	err = os.Setenv("AZURE_TENANT_ID", tokens["tenant_id"])
	if err != nil {
		return err
	}
	err = os.Setenv("AZURE_CLIENT_ID", tokens["client_id"])
	if err != nil {
		return err
	}
	err = os.Setenv("AZURE_CLIENT_SECRET", tokens["client_secret"])
	if err != nil {
		return err
	}
	return nil
}

func getAzureManagedClusterClient(cred *AzureProvider) (*armcontainerservice.ManagedClustersClient, error) {

	managedClustersClient, err := armcontainerservice.NewManagedClustersClient(cred.SubscriptionID, cred.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return managedClustersClient, nil
}

func getAzureResourceGroupsClient(cred *AzureProvider) (*armresources.ResourceGroupsClient, error) {

	resourceGroupClient, err := armresources.NewResourceGroupsClient(cred.SubscriptionID, cred.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return resourceGroupClient, nil
}

func (obj *AzureProvider) CreateResourceGroup(ctx context.Context) (*armresources.ResourceGroupsClientCreateOrUpdateResponse, error) {
	resourceGroupClient, err := getAzureResourceGroupsClient(obj)
	if err != nil {
		return nil, err
	}
	resourceGroup, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		obj.Config.ResourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(obj.Region),
		},
		nil)
	if err != nil {
		return nil, err
	}
	log.Printf("Created resource group: {%s}", *resourceGroup.Name)
	return &resourceGroup, nil
}

func (obj *AzureProvider) DeleteResourceGroup(ctx context.Context) error {
	resourceGroupClient, err := getAzureResourceGroupsClient(obj)
	if err != nil {
		return err
	}
	pollerResp, err := resourceGroupClient.BeginDelete(ctx, obj.Config.ResourceGroupName, nil)
	if err != nil {
		return err
	}
	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	log.Printf("Deleted resource group: {%s}", obj.Config.ResourceGroupName)
	return nil
}

func (obj *AzureProvider) CreateSubnet(ctx context.Context, subnetName string) (*armnetwork.Subnet, error) {
	subnetClient, err := armnetwork.NewSubnetsClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}

	parameters := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: to.Ptr("10.1.0.0/16"),
		},
	}

	pollerResponse, err := subnetClient.BeginCreateOrUpdate(ctx, obj.Config.ResourceGroupName, obj.Config.VirtualNetworkName, subnetName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	obj.Config.SubnetName = subnetName
	obj.Config.SubnetID = *resp.ID
	log.Printf("Created subnet: {%s}", *resp.Name)
	return &resp.Subnet, nil
}

func (obj *AzureProvider) DeleteSubnet(ctx context.Context, subnetName string) error {
	subnetClient, err := armnetwork.NewSubnetsClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}

	pollerResponse, err := subnetClient.BeginDelete(ctx, obj.Config.ResourceGroupName, obj.Config.VirtualNetworkName, subnetName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	log.Printf("Deleted subnet: {%s}", subnetName)
	return nil
}

func (obj *AzureProvider) CreatePublicIP(ctx context.Context, publicIPName string) (*armnetwork.PublicIPAddress, error) {
	publicIPAddressClient, err := armnetwork.NewPublicIPAddressesClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}

	parameters := armnetwork.PublicIPAddress{
		Location: to.Ptr(obj.Region),
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodStatic), // Static or Dynamic
		},
	}

	pollerResponse, err := publicIPAddressClient.BeginCreateOrUpdate(ctx, obj.Config.ResourceGroupName, publicIPName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	log.Printf("Created public IP address: {%s}", *resp.Name)
	return &resp.PublicIPAddress, err
}

func (obj *AzureProvider) DeleteAllPublicIP(ctx context.Context) error {
	for _, interfaceName := range obj.Config.InfoControlPlanes.PublicIPNames {
		if err := obj.DeletePublicIP(ctx, interfaceName); err != nil {
			return err
		}
	}
	for _, interfaceName := range obj.Config.InfoWorkerPlanes.PublicIPNames {
		if err := obj.DeletePublicIP(ctx, interfaceName); err != nil {
			return err
		}
	}

	if len(obj.Config.InfoDatabase.PublicIPName) != 0 {
		if err := obj.DeletePublicIP(ctx, obj.Config.InfoDatabase.PublicIPName); err != nil {
			return err
		}
	}
	if len(obj.Config.InfoLoadBalancer.PublicIPName) != 0 {
		if err := obj.DeletePublicIP(ctx, obj.Config.InfoLoadBalancer.PublicIPName); err != nil {
			return err
		}
	}
	log.Println("Deleted all Public IPs")
	return nil
}

func (obj *AzureProvider) DeletePublicIP(ctx context.Context, publicIPName string) error {
	publicIPAddressClient, err := armnetwork.NewPublicIPAddressesClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}

	pollerResponse, err := publicIPAddressClient.BeginDelete(ctx, obj.Config.ResourceGroupName, publicIPName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	log.Printf("Deleted the pubIP: {%s}", publicIPName)
	return nil
}

func (obj *AzureProvider) DeleteSSHKeyPair(ctx context.Context) error {
	sshClient, err := armcompute.NewSSHPublicKeysClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}
	_, err = sshClient.Delete(ctx, obj.Config.ResourceGroupName, obj.Config.SSHKeyName, nil)
	if err != nil {
		return err
	}

	log.Printf("Deleted the ssh: {%s}", obj.Config.SSHKeyName)
	return nil
}

func (obj *AzureProvider) UploadSSHKey(ctx context.Context) (err error) {
	sshClient, err := armcompute.NewSSHPublicKeysClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return
	}
	path := util.GetPath(util.CLUSTER_PATH, "azure", "ha", obj.ClusterName+" "+obj.Config.ResourceGroupName+" "+obj.Region)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return
	}

	keyPairToUpload, err := util.CreateSSHKeyPair("azure", obj.ClusterName+" "+obj.Config.ResourceGroupName+" "+obj.Region)
	if err != nil {
		return
	}

	_, err = sshClient.Create(ctx, obj.Config.ResourceGroupName, obj.ClusterName+"-ssh", armcompute.SSHPublicKeyResource{
		Location:   to.Ptr(obj.Region),
		Properties: &armcompute.SSHPublicKeyResourceProperties{PublicKey: to.Ptr(keyPairToUpload)},
	}, nil)
	obj.Config.SSHKeyName = obj.ClusterName + "-ssh"

	// ------- Setting the ssh configs only the public ips used will change
	obj.SSH_Payload.UserName = "azureuser"
	obj.SSH_Payload.PathPrivateKey = util.GetPath(util.SSH_PATH, "azure", "ha", obj.ClusterName+" "+obj.Config.ResourceGroupName+" "+obj.Region)
	obj.SSH_Payload.Output = ""
	obj.SSH_Payload.PublicIP = ""
	// ------

	return
}

func (obj *AzureProvider) CreateNetworkInterface(ctx context.Context, resourceName, nicName string, subnetID string, publicIPID string, networkSecurityGroupID string) (*armnetwork.Interface, error) {
	nicClient, err := armnetwork.NewInterfacesClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	parameters := armnetwork.Interface{
		Location: to.Ptr(obj.Region),
		Properties: &armnetwork.InterfacePropertiesFormat{
			IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
				{
					Name: to.Ptr(resourceName),
					Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
						PrivateIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
						Subnet: &armnetwork.Subnet{
							ID: to.Ptr(subnetID),
						},
						PublicIPAddress: &armnetwork.PublicIPAddress{
							ID: to.Ptr(publicIPID),
						},
					},
				},
			},
			NetworkSecurityGroup: &armnetwork.SecurityGroup{
				ID: to.Ptr(networkSecurityGroupID),
			},
		},
	}

	pollerResponse, err := nicClient.BeginCreateOrUpdate(ctx, obj.Config.ResourceGroupName, nicName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	log.Printf("Created network interface: {%s}", *resp.Name)
	return &resp.Interface, err
}

func (obj *AzureProvider) DeleteAllNetworkInterface(ctx context.Context) error {
	for _, interfaceName := range obj.Config.InfoControlPlanes.NetworkInterfaceNames {
		if err := obj.DeleteNetworkInterface(ctx, interfaceName); err != nil {
			return err
		}
	}
	for _, interfaceName := range obj.Config.InfoWorkerPlanes.NetworkInterfaceNames {
		if err := obj.DeleteNetworkInterface(ctx, interfaceName); err != nil {
			return err
		}
	}

	if len(obj.Config.InfoDatabase.NetworkInterfaceName) != 0 {
		if err := obj.DeleteNetworkInterface(ctx, obj.Config.InfoDatabase.NetworkInterfaceName); err != nil {
			return err
		}
	}
	if len(obj.Config.InfoLoadBalancer.NetworkInterfaceName) != 0 {
		if err := obj.DeleteNetworkInterface(ctx, obj.Config.InfoLoadBalancer.NetworkInterfaceName); err != nil {
			return err
		}
	}
	log.Println("Deleted all network interfaces")
	return nil
}

func (obj *AzureProvider) DeleteNetworkInterface(ctx context.Context, nicName string) error {
	nicClient, err := armnetwork.NewInterfacesClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}

	pollerResponse, err := nicClient.BeginDelete(ctx, obj.Config.ResourceGroupName, nicName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	log.Printf("Deleted the nic: {%s}", nicName)

	return nil
}

func (obj *AzureProvider) DeleteAllNSG(ctx context.Context) error {
	if len(obj.Config.InfoControlPlanes.NetworkSecurityGroupName) != 0 {
		if err := obj.DeleteNSG(ctx, obj.Config.InfoControlPlanes.NetworkSecurityGroupName); err != nil {
			return err
		}
	}
	if len(obj.Config.InfoWorkerPlanes.NetworkSecurityGroupName) != 0 {
		if err := obj.DeleteNSG(ctx, obj.Config.InfoWorkerPlanes.NetworkSecurityGroupName); err != nil {
			return err
		}
	}

	if len(obj.Config.InfoDatabase.NetworkSecurityGroupName) != 0 {
		if err := obj.DeleteNSG(ctx, obj.Config.InfoDatabase.NetworkSecurityGroupName); err != nil {
			return err
		}
	}
	if len(obj.Config.InfoLoadBalancer.NetworkSecurityGroupName) != 0 {
		if err := obj.DeleteNSG(ctx, obj.Config.InfoLoadBalancer.NetworkSecurityGroupName); err != nil {
			return err
		}
	}
	log.Println("Deleted all network security groups")
	return nil
}

func (obj *AzureProvider) DeleteNSG(ctx context.Context, nsgName string) error {
	nsgClient, err := armnetwork.NewSecurityGroupsClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}

	pollerResponse, err := nsgClient.BeginDelete(ctx, obj.Config.ResourceGroupName, nsgName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	log.Printf("Deleted the nsg: {%s}", nsgName)
	return nil
}

func (obj *AzureProvider) CreateNSG(ctx context.Context, nsgName string, securityRules []*armnetwork.SecurityRule) (*armnetwork.SecurityGroup, error) {
	nsgClient, err := armnetwork.NewSecurityGroupsClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}

	parameters := armnetwork.SecurityGroup{
		Location: to.Ptr(obj.Region),
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: securityRules,
		},
	}

	pollerResponse, err := nsgClient.BeginCreateOrUpdate(ctx, obj.Config.ResourceGroupName, nsgName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	log.Printf("Created network security group: {%s}", *resp.Name)
	return &resp.SecurityGroup, nil
}

func (obj *AzureProvider) DeleteVirtualNetwork(ctx context.Context) error {
	vnetClient, err := armnetwork.NewVirtualNetworksClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}

	pollerResponse, err := vnetClient.BeginDelete(ctx, obj.Config.ResourceGroupName, obj.Config.VirtualNetworkName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	log.Printf("Deleted virtual network {%s}", obj.Config.VirtualNetworkName)
	return nil
}

func (obj *AzureProvider) CreateVirtualNetwork(ctx context.Context, virtualNetworkName string) (*armnetwork.VirtualNetwork, error) {
	vnetClient, err := armnetwork.NewVirtualNetworksClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	parameters := armnetwork.VirtualNetwork{
		Location: to.Ptr(obj.Region),
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{
				AddressPrefixes: []*string{
					to.Ptr("10.1.0.0/16"), // example 10.1.0.0/16
				},
			},
		},
	}

	pollerResponse, err := vnetClient.BeginCreateOrUpdate(ctx, obj.Config.ResourceGroupName, virtualNetworkName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	// TODO: call the configWriter
	obj.Config.VirtualNetworkName = *resp.Name
	obj.Config.VirtualNetworkID = *resp.ID
	log.Printf("Created virtual network: {%s}", *resp.Name)
	return &resp.VirtualNetwork, nil
}

func (obj *AzureProvider) kubeconfigReader() ([]byte, error) {
	clusterDirName := obj.ClusterName + " " + obj.Config.ResourceGroupName + " " + obj.Region
	typeOfCluster := "managed"
	if obj.HACluster {
		typeOfCluster = "ha"
	}
	return os.ReadFile(util.GetPath(util.CLUSTER_PATH, "azure", typeOfCluster, clusterDirName, "config"))
}

func (p printer) Printer(isHA bool, operation int) {
	preFix := "export "
	if runtime.GOOS == "windows" {
		preFix = "$Env:"
	}
	switch operation {
	case 0:
		fmt.Printf("\n\033[33;40mTo use this cluster set this environment variable\033[0m\n\n")
		if isHA {
			fmt.Println(fmt.Sprintf("%sKUBECONFIG=\"%s\"\n", preFix, util.GetPath(util.CLUSTER_PATH, "azure", "ha", p.ClusterName+" "+p.ResourceName+" "+p.Region, "config")))
		} else {
			fmt.Println(fmt.Sprintf("%sKUBECONFIG=\"%s\"\n", preFix, util.GetPath(util.CLUSTER_PATH, "azure", "managed", p.ClusterName+" "+p.ResourceName+" "+p.Region, "config")))
		}
	case 1:
		fmt.Printf("\n\033[33;40mUse the following command to unset KUBECONFIG\033[0m\n\n")
		if runtime.GOOS == "windows" {
			fmt.Println(fmt.Sprintf("%sKUBECONFIG=\"\"\n", preFix))
		} else {
			fmt.Println("unset KUBECONFIG")
		}
	}
	fmt.Println()
}

func (obj *AzureProvider) CreateVM(ctx context.Context, vmName, networkInterfaceID, diskName, script string) (*armcompute.VirtualMachine, error) {
	vmClient, err := armcompute.NewVirtualMachinesClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}

	//require ssh key for authentication on linux
	sshPublicKeyPath := util.GetPath(util.OTHER_PATH, "azure", "ha", obj.ClusterName+" "+obj.Config.ResourceGroupName+" "+obj.Region, "keypair.pub")
	var sshBytes []byte
	_, err = os.Stat(sshPublicKeyPath)
	if err != nil {
		return nil, err
	}

	sshBytes, err = os.ReadFile(sshPublicKeyPath)
	if err != nil {
		return nil, err
	}

	parameters := armcompute.VirtualMachine{
		Location: to.Ptr(obj.Region),
		Identity: &armcompute.VirtualMachineIdentity{
			Type: to.Ptr(armcompute.ResourceIdentityTypeNone),
		},
		Properties: &armcompute.VirtualMachineProperties{
			StorageProfile: &armcompute.StorageProfile{
				ImageReference: &armcompute.ImageReference{
					Offer:     to.Ptr("0001-com-ubuntu-server-jammy"),
					Publisher: to.Ptr("Canonical"),
					SKU:       to.Ptr("22_04-lts-gen2"),
					Version:   to.Ptr("latest"),
				},
				OSDisk: &armcompute.OSDisk{
					Name:         to.Ptr(diskName),
					CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesFromImage),
					Caching:      to.Ptr(armcompute.CachingTypesReadWrite),
					ManagedDisk: &armcompute.ManagedDiskParameters{
						StorageAccountType: to.Ptr(armcompute.StorageAccountTypesStandardLRS), // OSDisk type Standard/Premium HDD/SSD
					},
					//DiskSizeGB: to.Ptr[int32](100), // default 127G
				},
			},
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: to.Ptr(armcompute.VirtualMachineSizeTypes(obj.Spec.Disk)), // VM size include vCPUs,RAM,Data Disks,Temp storage.
			},
			OSProfile: &armcompute.OSProfile{ //
				ComputerName:  to.Ptr(vmName),
				AdminUsername: to.Ptr("azureuser"),
				CustomData:    to.Ptr(base64.StdEncoding.EncodeToString([]byte(script))),
				LinuxConfiguration: &armcompute.LinuxConfiguration{
					DisablePasswordAuthentication: to.Ptr(true),
					SSH: &armcompute.SSHConfiguration{
						PublicKeys: []*armcompute.SSHPublicKey{
							{
								Path:    to.Ptr(string("/home/azureuser/.ssh/authorized_keys")),
								KeyData: to.Ptr(string(sshBytes)),
							},
						},
					},
				},
			},
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: to.Ptr(networkInterfaceID),
					},
				},
			},
		},
	}

	pollerResponse, err := vmClient.BeginCreateOrUpdate(ctx, obj.Config.ResourceGroupName, vmName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	log.Printf("Created network virtual machine: {%s}", *resp.Name)
	return &resp.VirtualMachine, nil
}

func (obj *AzureProvider) DeleteAllVMs(ctx context.Context) error {
	for _, instanceName := range obj.Config.InfoControlPlanes.Names {
		if err := obj.DeleteVM(ctx, instanceName); err != nil {
			return err
		}
	}
	for _, instanceName := range obj.Config.InfoWorkerPlanes.Names {
		if err := obj.DeleteVM(ctx, instanceName); err != nil {
			return err
		}
	}

	if len(obj.Config.InfoDatabase.Name) != 0 {
		if err := obj.DeleteVM(ctx, obj.Config.InfoDatabase.Name); err != nil {
			return err
		}
	}
	if len(obj.Config.InfoLoadBalancer.Name) != 0 {
		if err := obj.DeleteVM(ctx, obj.Config.InfoLoadBalancer.Name); err != nil {
			return err
		}
	}
	log.Println("Deleted all virtual machines")
	return nil
}

func (obj *AzureProvider) DeleteVM(ctx context.Context, vmName string) error {
	vmClient, err := armcompute.NewVirtualMachinesClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}

	pollerResponse, err := vmClient.BeginDelete(ctx, obj.Config.ResourceGroupName, vmName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	log.Printf("Deleted the vm: {%s}", vmName)
	return nil
}

func (obj *AzureProvider) DeleteAllDisks(ctx context.Context) error {
	for _, diskName := range obj.Config.InfoControlPlanes.DiskNames {
		if err := obj.DeleteDisk(ctx, diskName); err != nil {
			return err
		}
	}
	for _, diskName := range obj.Config.InfoWorkerPlanes.DiskNames {
		if err := obj.DeleteDisk(ctx, diskName); err != nil {
			return err
		}
	}

	if len(obj.Config.InfoDatabase.DiskName) != 0 {
		if err := obj.DeleteDisk(ctx, obj.Config.InfoDatabase.DiskName); err != nil {
			return err
		}
	}
	if len(obj.Config.InfoLoadBalancer.DiskName) != 0 {
		if err := obj.DeleteDisk(ctx, obj.Config.InfoLoadBalancer.DiskName); err != nil {
			return err
		}
	}
	log.Println("Deleted all disks")
	return nil
}

func (obj *AzureProvider) DeleteDisk(ctx context.Context, diskName string) error {
	diskClient, err := armcompute.NewDisksClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}

	pollerResponse, err := diskClient.BeginDelete(ctx, obj.Config.ResourceGroupName, diskName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	log.Printf("Deleted disk: {%s}", diskName)
	return nil
}

// SaveKubeconfig stores the kubeconfig to state management file
func (obj *AzureProvider) SaveKubeconfig(kubeconfig string) error {
	folderName := obj.ClusterName + " " + obj.Config.ResourceGroupName + " " + obj.Region
	kind := "managed"
	if obj.HACluster {
		kind = "ha"
	}
	err := os.MkdirAll(util.GetPath(util.CLUSTER_PATH, "azure", kind, folderName), 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	_, err = os.Create(util.GetPath(util.CLUSTER_PATH, "azure", kind, folderName, "config"))
	if err != nil && !os.IsExist(err) {
		return err
	}

	file, err := os.OpenFile(util.GetPath(util.CLUSTER_PATH, "azure", kind, folderName, "config"), os.O_WRONLY, 0750)
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
