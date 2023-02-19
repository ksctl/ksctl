package azure

import (
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
}

type AzureStateVMs struct {
	Names                    []string `json:"names"`
	NetworkSecurityGroupName string   `json:"network_security_group_name"`
	DiskNames                []string `json:"disk_names"`
	PublicIPNames            []string `json:"public_ip_names"`
	NetworkInterfaceNames    []string `json:"network_interface_names"`
}
type AzureStateVM struct {
	Name                     string `json:"name"`
	NetworkSecurityGroupName string `json:"network_security_group_name"`
	DiskName                 string `json:"disk_name"`
	PublicIPName             string `json:"public_ip_name"`
	NetworkInterfaceName     string `json:"network_interface_name"`
}

type AzureStateCluster struct {
	ClusterName        string        `json:"cluster_name"`
	ResourceGroupName  string        `json:"resource_group_name"`
	SSHKeyName         string        `json:"ssh_key_name"`
	SubnetName         string        `json:"subnet_name"`
	VirtualNetworkName string        `json:"virtual_network_name"`
	InfoControlPlanes  AzureStateVMs `json:"info_control_planes"`
	InfoWorkerPlanes   AzureStateVMs `json:"info_worker_planes"`
	InfoDatabase       AzureStateVM  `json:"info_database"`
	InfoLoadBalancer   AzureStateVM  `json:"info_load_balancer"`
}

type AzureInfra interface {

	// azure resources
	CreateResourceGroup(context.Context) error
	DeleteResourceGroup(context.Context) error
	DeleteDisk(context.Context, string) error
	DeleteSubnet(context.Context) error
	CreateSubnet(context.Context) (*armnetwork.Subnet, error)
	CreateVM(context.Context, string, string, string) (*armcompute.VirtualMachine, error)
	DeleteVM(context.Context, string) error
	CreateVirtualNetwork(context.Context) (*armnetwork.VirtualNetwork, error)
	DeleteVirtualNetwork(context.Context) error
	CreateNSG(context.Context, string) (*armnetwork.SecurityGroup, error)
	DeleteNSG(context.Context, string) error
	DeleteNetworkInterface(context.Context, string) error
	CreateNetworkInterface(context.Context, string, string, string, string, string) (*armnetwork.Interface, error)
	DeletePublicIP(context.Context, string) error
	CreatePublicIP(context.Context, string) (*armnetwork.PublicIPAddress, error)

	// state file managemenet
	ConfigWriterManagedClusteName() error
	ConfigWriterManagedResourceName() error
	ConfigReaderManaged() error

	// kubeconfig file
	kubeconfigWriter(string) error
	kubeconfigReader() ([]byte, error)
}

func (config *AzureProvider) ConfigWriterManagedClusteName() error {
	config.Config.ClusterName = config.ClusterName
	return util.SaveState(config.Config, "azure", config.ClusterName+" "+config.Config.ResourceGroupName+" "+config.Region)
}

func (config *AzureProvider) ConfigWriterManagedResourceName() error {
	return util.SaveState(config.Config, "azure", config.ClusterName+" "+config.Config.ResourceGroupName+" "+config.Region)
}

func (config *AzureProvider) ConfigReaderManaged() error {
	data, err := util.GetState("azure", config.ClusterName+" "+config.Config.ResourceGroupName+" "+config.Region)
	if err != nil {
		return err
	}
	// populating the state data
	config.Config.ClusterName = data["cluster_name"]
	config.Config.ResourceGroupName = data["resource_group_name"]
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

	return nil
}

func (obj *AzureProvider) CreateSubnet(ctx context.Context) (*armnetwork.Subnet, error) {
	subnetClient, err := armnetwork.NewSubnetsClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}

	parameters := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: to.Ptr("10.1.0.0/16"),
		},
	}

	pollerResponse, err := subnetClient.BeginCreateOrUpdate(ctx, obj.Config.ResourceGroupName, obj.Config.VirtualNetworkName, obj.Config.SubnetName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &resp.Subnet, nil
}

func (obj *AzureProvider) DeleteSubnet(ctx context.Context) error {
	subnetClient, err := armnetwork.NewSubnetsClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}

	pollerResponse, err := subnetClient.BeginDelete(ctx, obj.Config.ResourceGroupName, obj.Config.VirtualNetworkName, obj.Config.SubnetName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

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
	return &resp.PublicIPAddress, err
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
	return nil
}

func (obj *AzureProvider) CreateNetworkInterface(ctx context.Context, resourceName, nicName string, subnetID string, publicIPID string, networkSecurityGroupID string) (*armnetwork.Interface, error) {
	nicClient, err := armnetwork.NewInterfacesClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	parameters := armnetwork.Interface{
		Location: to.Ptr(obj.Region),
		Properties: &armnetwork.InterfacePropertiesFormat{
			//NetworkSecurityGroup:
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

	return &resp.Interface, err
}

func (obj *AzureProvider) DeleteNetworkInterface(ctx context.Context, nicName string) error {
	nicClient, err := armnetwork.NewInterfacesClient(obj.Config.ResourceGroupName, obj.AzureTokenCred, nil)
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
	return nil
}

func (obj *AzureProvider) CreateNSG(ctx context.Context, nsgName string) (*armnetwork.SecurityGroup, error) {
	nsgClient, err := armnetwork.NewSecurityGroupsClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}

	parameters := armnetwork.SecurityGroup{
		Location: to.Ptr(obj.Region),
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: []*armnetwork.SecurityRule{
				// Windows connection to virtual machine needs to open port 3389,RDP
				// inbound
				{
					Name: to.Ptr("sample_inbound_22"), //
					Properties: &armnetwork.SecurityRulePropertiesFormat{
						SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
						SourcePortRange:          to.Ptr("*"),
						DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
						DestinationPortRange:     to.Ptr("22"),
						Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
						Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
						Priority:                 to.Ptr[int32](100),
						Description:              to.Ptr("sample network security group inbound port 22"),
						Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
					},
				},
				// outbound
				{
					Name: to.Ptr("sample_outbound_22"), //
					Properties: &armnetwork.SecurityRulePropertiesFormat{
						SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
						SourcePortRange:          to.Ptr("*"),
						DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
						DestinationPortRange:     to.Ptr("22"),
						Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
						Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
						Priority:                 to.Ptr[int32](100),
						Description:              to.Ptr("sample network security group outbound port 22"),
						Direction:                to.Ptr(armnetwork.SecurityRuleDirectionOutbound),
					},
				},
			},
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

	return nil
}

func (obj *AzureProvider) CreateVirtualNetwork(ctx context.Context) (*armnetwork.VirtualNetwork, error) {
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
			//Subnets: []*armnetwork.Subnet{
			//	{
			//		Name: to.Ptr(subnetName+"3"),
			//		Properties: &armnetwork.SubnetPropertiesFormat{
			//			AddressPrefix: to.Ptr("10.1.0.0/24"),
			//		},
			//	},
			//},
		},
	}

	pollerResponse, err := vnetClient.BeginCreateOrUpdate(ctx, obj.Config.ResourceGroupName, obj.Config.VirtualNetworkName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	// TODO: call the configWriter

	return &resp.VirtualNetwork, nil
}

func (obj *AzureProvider) kubeconfigWriter(kubeconfig string) error {
	clusterDirName := obj.ClusterName + " " + obj.Config.ResourceGroupName + " " + obj.Region
	typeOfCluster := "managed"
	if obj.HACluster {
		typeOfCluster = "ha"
	}
	err := os.WriteFile(util.GetPath(util.CLUSTER_PATH, "azure", typeOfCluster, clusterDirName, "config"), []byte(kubeconfig), 0644)
	if err != nil {
		return err
	}
	log.Println("💾 configuration")
	return nil
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

func (obj *AzureProvider) CreateVM(ctx context.Context, vmName, networkInterfaceID, diskName string) (*armcompute.VirtualMachine, error) {
	vmClient, err := armcompute.NewVirtualMachinesClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}

	//require ssh key for authentication on linux
	//sshPublicKeyPath := "/home/user/.ssh/id_rsa.pub"
	//var sshBytes []byte
	//_,err := os.Stat(sshPublicKeyPath)
	//if err == nil {
	//	sshBytes,err = ioutil.ReadFile(sshPublicKeyPath)
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	parameters := armcompute.VirtualMachine{
		Location: to.Ptr(obj.Region),
		Identity: &armcompute.VirtualMachineIdentity{
			Type: to.Ptr(armcompute.ResourceIdentityTypeNone),
		},
		Properties: &armcompute.VirtualMachineProperties{
			StorageProfile: &armcompute.StorageProfile{
				ImageReference: &armcompute.ImageReference{
					// search image reference
					// az vm image list --output table
					Offer:     to.Ptr("WindowsServer"),
					Publisher: to.Ptr("MicrosoftWindowsServer"),
					SKU:       to.Ptr("2019-Datacenter"),
					Version:   to.Ptr("latest"),
					//require ssh key for authentication on linux
					//Offer:     to.Ptr("UbuntuServer"),
					//Publisher: to.Ptr("Canonical"),
					//SKU:       to.Ptr("18.04-LTS"),
					//Version:   to.Ptr("latest"),
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
				VMSize: to.Ptr(armcompute.VirtualMachineSizeTypes("Standard_F2s")), // VM size include vCPUs,RAM,Data Disks,Temp storage.
			},
			OSProfile: &armcompute.OSProfile{ //
				ComputerName:  to.Ptr("sample-compute"),
				AdminUsername: to.Ptr("sample-user"),
				AdminPassword: to.Ptr("Password01!@#"),
				//require ssh key for authentication on linux
				//LinuxConfiguration: &armcompute.LinuxConfiguration{
				//	DisablePasswordAuthentication: to.Ptr(true),
				//	SSH: &armcompute.SSHConfiguration{
				//		PublicKeys: []*armcompute.SSHPublicKey{
				//			{
				//				Path:    to.Ptr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", "sample-user")),
				//				KeyData: to.Ptr(string(sshBytes)),
				//			},
				//		},
				//	},
				//},
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

	return &resp.VirtualMachine, nil
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
	return nil
}
