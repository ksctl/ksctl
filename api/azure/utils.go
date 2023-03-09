package azure

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"runtime"

	log "github.com/kubesimplify/ksctl/api/logger"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	util "github.com/kubesimplify/ksctl/api/utils"
	"golang.org/x/net/context"
)

type AzureOperations interface {
	CreateCluster(log.Logger) error
	DeleteCluster(log.Logger) error

	AddMoreWorkerNodes(log.Logger) error
	DeleteSomeWorkerNodes(log.Logger) error
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
	CreateResourceGroup(context.Context, log.Logger) error
	DeleteResourceGroup(context.Context, log.Logger) error
	DeleteDisk(context.Context, log.Logger, string) error
	DeleteSubnet(context.Context, log.Logger) error
	CreateSubnet(context.Context, log.Logger, string) (*armnetwork.Subnet, error)
	CreateVM(context.Context, log.Logger, string, string, string) (*armcompute.VirtualMachine, error)
	DeleteVM(context.Context, log.Logger, string) error
	CreateVirtualNetwork(context.Context, log.Logger, string) (*armnetwork.VirtualNetwork, error)
	DeleteVirtualNetwork(context.Context, log.Logger) error
	CreateNSG(context.Context, log.Logger, string, []*armnetwork.SecurityRule) (*armnetwork.SecurityGroup, error)
	DeleteNSG(context.Context, log.Logger, string) error
	DeleteNetworkInterface(context.Context, log.Logger, string) error
	CreateNetworkInterface(context.Context, log.Logger, string, string, string, string, string) (*armnetwork.Interface, error)
	DeletePublicIP(context.Context, log.Logger, string) error
	CreatePublicIP(context.Context, log.Logger, string) (*armnetwork.PublicIPAddress, error)

	UploadSSHKey(context.Context, log.Logger) (err error)
	DeleteSSHKey(context.Context, log.Logger) error

	// state file managemenet
	ConfigReader() error
	ConfigWriter() error

	// kubeconfig file
	kubeconfigWriter(string) error
	kubeconfigReader() ([]byte, error)
}

func isValidNodeSize(disk string) bool {
	validDisks := []string{
		"Standard_B1ms", "Standard_B1s", "Standard_B2ms", "Standard_B1ls", "Standard_B2s", "Standard_B4ms",
		"Standard_B8ms", "Standard_B12ms", "Standard_B16ms", "Standard_B20ms", "Standard_E2_v4", "Standard_E4_v4",
		"Standard_E8_v4", "Standard_E16_v4", "Standard_E20_v4", "Standard_E32_v4", "Standard_E2d_v4", "Standard_E4d_v4",
		"Standard_E8d_v4", "Standard_E16d_v4", "Standard_E20d_v4", "Standard_E32d_v4", "Standard_E2s_v4", "Standard_E4-2s_v4",
		"Standard_E4s_v4", "Standard_E8-2s_v4", "Standard_E8-4s_v4", "Standard_E8s_v4", "Standard_E16-4s_v4",
		"Standard_E16-8s_v4", "Standard_E16s_v4", "Standard_E20s_v4", "Standard_E32-8s_v4", "Standard_E32-16s_v4", "Standard_E32s_v4",
		"Standard_E2ds_v4", "Standard_E4-2ds_v4", "Standard_E4ds_v4", "Standard_E8-2ds_v4", "Standard_E8-4ds_v4", "Standard_E8ds_v4",
		"Standard_E16-4ds_v4",
		"Standard_E16-8ds_v4",
		"Standard_E16ds_v4",
		"Standard_E20ds_v4",
		"Standard_E32-8ds_v4",
		"Standard_E32-16ds_v4",
		"Standard_E32ds_v4",
		"Standard_D2d_v4",
		"Standard_D4d_v4",
		"Standard_D8d_v4",
		"Standard_D16d_v4",
		"Standard_D32d_v4",
		"Standard_D48d_v4",
		"Standard_D64d_v4",
		"Standard_D2_v4",
		"Standard_D4_v4",
		"Standard_D8_v4",
		"Standard_D16_v4",
		"Standard_D32_v4",
		"Standard_D48_v4",
		"Standard_D64_v4",
		"Standard_D2ds_v4",
		"Standard_D4ds_v4",
		"Standard_D8ds_v4",
		"Standard_D16ds_v4",
		"Standard_D32ds_v4",
		"Standard_D48ds_v4",
		"Standard_D64ds_v4",
		"Standard_D2s_v4",
		"Standard_D4s_v4",
		"Standard_D8s_v4",
		"Standard_D16s_v4",
		"Standard_D32s_v4",
		"Standard_D48s_v4",
		"Standard_D64s_v4",
		"Standard_D1_v2",
		"Standard_D2_v2",
		"Standard_D3_v2",
		"Standard_D4_v2",
		"Standard_D5_v2",
		"Standard_D11_v2",
		"Standard_D12_v2",
		"Standard_D13_v2",
		"Standard_D14_v2",
		"Standard_D15_v2",
		"Standard_D2_v2_Promo",
		"Standard_D3_v2_Promo",
		"Standard_D4_v2_Promo",
		"Standard_D5_v2_Promo",
		"Standard_D11_v2_Promo",
		"Standard_D12_v2_Promo",
		"Standard_D13_v2_Promo",
		"Standard_D14_v2_Promo",
		"Standard_F1",
		"Standard_F2",
		"Standard_F4",
		"Standard_F8",
		"Standard_F16",
		"Standard_DS1_v2",
		"Standard_DS2_v2",
		"Standard_DS3_v2",
		"Standard_DS4_v2",
		"Standard_DS5_v2",
		"Standard_DS11-1_v2",
		"Standard_DS11_v2",
		"Standard_DS12-1_v2",
		"Standard_DS12-2_v2",
		"Standard_DS12_v2",
		"Standard_DS13-2_v2",
		"Standard_DS13-4_v2",
		"Standard_DS13_v2",
		"Standard_DS14-4_v2",
		"Standard_DS14-8_v2",
		"Standard_DS14_v2",
		"Standard_DS15_v2",
		"Standard_DS2_v2_Promo",
		"Standard_DS3_v2_Promo",
		"Standard_DS4_v2_Promo",
		"Standard_DS5_v2_Promo",
		"Standard_DS11_v2_Promo",
		"Standard_DS12_v2_Promo",
		"Standard_DS13_v2_Promo",
		"Standard_DS14_v2_Promo",
		"Standard_F1s",
		"Standard_F2s",
		"Standard_F4s",
		"Standard_F8s",
		"Standard_F16s",
		"Standard_A1_v2",
		"Standard_A2m_v2",
		"Standard_A2_v2",
		"Standard_A4m_v2",
		"Standard_A4_v2",
		"Standard_A8m_v2",
		"Standard_A8_v2",
		"Standard_D2_v3",
		"Standard_D4_v3",
		"Standard_D8_v3",
		"Standard_D16_v3",
		"Standard_D32_v3",
		"Standard_D48_v3",
		"Standard_D64_v3",
		"Standard_D2s_v3",
		"Standard_D4s_v3",
		"Standard_D8s_v3",
		"Standard_D16s_v3",
		"Standard_D32s_v3",
		"Standard_D48s_v3",
		"Standard_D64s_v3",
		"Standard_E2_v3",
		"Standard_E4_v3",
		"Standard_E8_v3",
		"Standard_E16_v3",
		"Standard_E20_v3",
		"Standard_E32_v3",
		"Standard_E2s_v3",
		"Standard_E4-2s_v3",
		"Standard_E4s_v3",
		"Standard_E8-2s_v3",
		"Standard_E8-4s_v3",
		"Standard_E8s_v3",
		"Standard_E16-4s_v3",
		"Standard_E16-8s_v3",
		"Standard_E16s_v3",
		"Standard_E20s_v3",
		"Standard_E32-8s_v3",
		"Standard_E32-16s_v3",
		"Standard_E32s_v3",
		"Standard_F2s_v2",
		"Standard_F4s_v2",
		"Standard_F8s_v2",
		"Standard_F16s_v2",
		"Standard_F32s_v2",
		"Standard_F48s_v2",
		"Standard_F64s_v2",
		"Standard_F72s_v2",
		"Standard_D2a_v4",
		"Standard_D4a_v4",
		"Standard_D8a_v4",
		"Standard_D16a_v4",
		"Standard_D32a_v4",
		"Standard_D48a_v4",
		"Standard_D64a_v4",
		"Standard_D96a_v4",
		"Standard_D2as_v4",
		"Standard_D4as_v4",
		"Standard_D8as_v4",
		"Standard_D16as_v4",
		"Standard_D32as_v4",
		"Standard_D48as_v4",
		"Standard_D64as_v4",
		"Standard_D96as_v4",
		"Standard_E2a_v4",
		"Standard_E4a_v4",
		"Standard_E8a_v4",
		"Standard_E16a_v4",
		"Standard_E20a_v4",
		"Standard_E32a_v4",
		"Standard_E48a_v4",
		"Standard_E64a_v4",
		"Standard_E96a_v4",
		"Standard_E2as_v4",
		"Standard_E4-2as_v4",
		"Standard_E4as_v4",
		"Standard_E8-2as_v4",
		"Standard_E8-4as_v4",
		"Standard_E8as_v4",
		"Standard_E16-4as_v4",
		"Standard_E16-8as_v4",
		"Standard_E16as_v4",
		"Standard_E20as_v4",
		"Standard_E32-8as_v4",
		"Standard_E32-16as_v4",
		"Standard_E32as_v4",
		"Standard_E48as_v4",
		"Standard_E64-16as_v4",
		"Standard_E64-32as_v4",
		"Standard_E64as_v4",
		"Standard_E96-24as_v4",
		"Standard_E96-48as_v4",
		"Standard_E96as_v4",
		"Standard_D2as_v5",
		"Standard_D4as_v5",
		"Standard_D8as_v5",
		"Standard_D16as_v5",
		"Standard_D32as_v5",
		"Standard_D48as_v5",
		"Standard_D64as_v5",
		"Standard_D96as_v5",
		"Standard_E2as_v5",
		"Standard_E4-2as_v5",
		"Standard_E4as_v5",
		"Standard_E8-2as_v5",
		"Standard_E8-4as_v5",
		"Standard_E8as_v5",
		"Standard_E16-4as_v5",
		"Standard_E16-8as_v5",
		"Standard_E16as_v5",
		"Standard_E20as_v5",
		"Standard_E32-8as_v5",
		"Standard_E32-16as_v5",
		"Standard_E32as_v5",
		"Standard_E48as_v5",
		"Standard_E64-16as_v5",
		"Standard_E64-32as_v5",
		"Standard_E64as_v5",
		"Standard_E96-24as_v5",
		"Standard_E96-48as_v5",
		"Standard_E96as_v5",
		"Standard_E112ias_v5",
		"Standard_D2ads_v5",
		"Standard_D4ads_v5",
		"Standard_D8ads_v5",
		"Standard_D16ads_v5",
		"Standard_D32ads_v5",
		"Standard_D48ads_v5",
		"Standard_D64ads_v5",
		"Standard_D96ads_v5",
		"Standard_E2ads_v5",
		"Standard_E4-2ads_v5",
		"Standard_E4ads_v5",
		"Standard_E8-2ads_v5",
		"Standard_E8-4ads_v5",
		"Standard_E8ads_v5",
		"Standard_E16-4ads_v5",
		"Standard_E16-8ads_v5",
		"Standard_E16ads_v5",
		"Standard_E20ads_v5",
		"Standard_E32-8ads_v5",
		"Standard_E32-16ads_v5",
		"Standard_E32ads_v5",
		"Standard_E48ads_v5",
		"Standard_E64-16ads_v5",
		"Standard_E64-32ads_v5",
		"Standard_E64ads_v5",
		"Standard_E96-24ads_v5",
		"Standard_E96-48ads_v5",
		"Standard_E96ads_v5",
		"Standard_E112iads_v5",
		"Standard_E48_v4",
		"Standard_E64_v4",
		"Standard_E48d_v4",
		"Standard_E64d_v4",
		"Standard_E48s_v4",
		"Standard_E64-16s_v4",
		"Standard_E64-32s_v4",
		"Standard_E64s_v4",
		"Standard_E80is_v4",
		"Standard_E48ds_v4",
		"Standard_E64-16ds_v4",
		"Standard_E64-32ds_v4",
		"Standard_E64ds_v4",
		"Standard_E80ids_v4",
		"Standard_E48_v3",
		"Standard_E64_v3",
		"Standard_E48s_v3",
		"Standard_E64-16s_v3",
		"Standard_E64-32s_v3",
		"Standard_E64s_v3",
		"Standard_A0",
		"Standard_A1",
		"Standard_A2",
		"Standard_A3",
		"Standard_A5",
		"Standard_A4",
		"Standard_A6",
		"Standard_A7",
		"Basic_A0",
		"Basic_A1",
		"Basic_A2",
		"Basic_A3",
		"Basic_A4",
		"Standard_L8as_v3",
		"Standard_L16as_v3",
		"Standard_L32as_v3",
		"Standard_L48as_v3",
		"Standard_L64as_v3",
		"Standard_L80as_v3",
		"Standard_NC4as_T4_v3",
		"Standard_NC8as_T4_v3",
		"Standard_NC16as_T4_v3",
		"Standard_NC64as_T4_v3",
		"Standard_M64",
		"Standard_M64m",
		"Standard_M128",
		"Standard_M128m",
		"Standard_M8-2ms",
		"Standard_M8-4ms",
		"Standard_M8ms",
		"Standard_M16-4ms",
		"Standard_M16-8ms",
		"Standard_M16ms",
		"Standard_M32-8ms",
		"Standard_M32-16ms",
		"Standard_M32ls",
		"Standard_M32ms",
		"Standard_M32ts",
		"Standard_M64-16ms",
		"Standard_M64-32ms",
		"Standard_M64ls",
		"Standard_M64ms",
		"Standard_M64s",
		"Standard_M128-32ms",
		"Standard_M128-64ms",
		"Standard_M128ms",
		"Standard_M128s",
		"Standard_M32ms_v2",
		"Standard_M64ms_v2",
		"Standard_M64s_v2",
		"Standard_M128ms_v2",
		"Standard_M128s_v2",
		"Standard_M192ims_v2",
		"Standard_M192is_v2",
		"Standard_M32dms_v2",
		"Standard_M64dms_v2",
		"Standard_M64ds_v2",
		"Standard_M128dms_v2",
		"Standard_M128ds_v2",
		"Standard_M192idms_v2",
		"Standard_M192ids_v2",
		"Standard_D1",
		"Standard_D2",
		"Standard_D3",
		"Standard_D4",
		"Standard_D11",
		"Standard_D12",
		"Standard_D13",
		"Standard_D14",
		"Standard_DS1",
		"Standard_DS2",
		"Standard_DS3",
		"Standard_DS4",
		"Standard_DS11",
		"Standard_DS12",
		"Standard_DS13",
		"Standard_DS14",
		"Standard_DC8_v2",
		"Standard_DC1s_v2",
		"Standard_DC2s_v2",
		"Standard_DC4s_v2",
		"Standard_D2ds_v5",
		"Standard_D4ds_v5",
		"Standard_D8ds_v5",
		"Standard_D16ds_v5",
		"Standard_D32ds_v5",
		"Standard_D48ds_v5",
		"Standard_D64ds_v5",
		"Standard_D96ds_v5",
		"Standard_D2d_v5",
		"Standard_D4d_v5",
		"Standard_D8d_v5",
		"Standard_D16d_v5",
		"Standard_D32d_v5",
		"Standard_D48d_v5",
		"Standard_D64d_v5",
		"Standard_D96d_v5",
		"Standard_D2s_v5",
		"Standard_D4s_v5",
		"Standard_D8s_v5",
		"Standard_D16s_v5",
		"Standard_D32s_v5",
		"Standard_D48s_v5",
		"Standard_D64s_v5",
		"Standard_D96s_v5",
		"Standard_D2_v5",
		"Standard_D4_v5",
		"Standard_D8_v5",
		"Standard_D16_v5",
		"Standard_D32_v5",
		"Standard_D48_v5",
		"Standard_D64_v5",
		"Standard_D96_v5",
		"Standard_E2ds_v5",
		"Standard_E4-2ds_v5",
		"Standard_E4ds_v5",
		"Standard_E8-2ds_v5",
		"Standard_E8-4ds_v5",
		"Standard_E8ds_v5",
		"Standard_E16-4ds_v5",
		"Standard_E16-8ds_v5",
		"Standard_E16ds_v5",
		"Standard_E20ds_v5",
		"Standard_E32-8ds_v5",
		"Standard_E32-16ds_v5",
		"Standard_E32ds_v5",
		"Standard_E48ds_v5",
		"Standard_E64-16ds_v5",
		"Standard_E64-32ds_v5",
		"Standard_E64ds_v5",
		"Standard_E96-24ds_v5",
		"Standard_E96-48ds_v5",
		"Standard_E96ds_v5",
		"Standard_E104ids_v5",
		"Standard_E2d_v5",
		"Standard_E4d_v5",
		"Standard_E8d_v5",
		"Standard_E16d_v5",
		"Standard_E20d_v5",
		"Standard_E32d_v5",
		"Standard_E48d_v5",
		"Standard_E64d_v5",
		"Standard_E96d_v5",
		"Standard_E104id_v5",
		"Standard_E2s_v5",
		"Standard_E4-2s_v5",
		"Standard_E4s_v5",
		"Standard_E8-2s_v5",
		"Standard_E8-4s_v5",
		"Standard_E8s_v5",
		"Standard_E16-4s_v5",
		"Standard_E16-8s_v5",
		"Standard_E16s_v5",
		"Standard_E20s_v5",
		"Standard_E32-8s_v5",
		"Standard_E32-16s_v5",
		"Standard_E32s_v5",
		"Standard_E48s_v5",
		"Standard_E64-16s_v5",
		"Standard_E64-32s_v5",
		"Standard_E64s_v5",
		"Standard_E96-24s_v5",
		"Standard_E96-48s_v5",
		"Standard_E96s_v5",
		"Standard_E104is_v5",
		"Standard_E2_v5",
		"Standard_E4_v5",
		"Standard_E8_v5",
		"Standard_E16_v5",
		"Standard_E20_v5",
		"Standard_E32_v5",
		"Standard_E48_v5",
		"Standard_E64_v5",
		"Standard_E96_v5",
		"Standard_E104i_v5",
		"Standard_E2bs_v5",
		"Standard_E4bs_v5",
		"Standard_E8bs_v5",
		"Standard_E16bs_v5",
		"Standard_E32bs_v5",
		"Standard_E48bs_v5",
		"Standard_E64bs_v5",
		"Standard_E2bds_v5",
		"Standard_E4bds_v5",
		"Standard_E8bds_v5",
		"Standard_E16bds_v5",
		"Standard_E32bds_v5",
		"Standard_E48bds_v5",
		"Standard_E64bds_v5",
		"Standard_D2ls_v5",
		"Standard_D4ls_v5",
		"Standard_D8ls_v5",
		"Standard_D16ls_v5",
		"Standard_D32ls_v5",
		"Standard_D48ls_v5",
		"Standard_D64ls_v5",
		"Standard_D96ls_v5",
		"Standard_D2lds_v5",
		"Standard_D4lds_v5",
		"Standard_D8lds_v5",
		"Standard_D16lds_v5",
		"Standard_D32lds_v5",
		"Standard_D48lds_v5",
		"Standard_D64lds_v5",
		"Standard_D96lds_v5",
		"Standard_L8s_v2",
		"Standard_L16s_v2",
		"Standard_L32s_v2",
		"Standard_L48s_v2",
		"Standard_L64s_v2",
		"Standard_L80s_v2",
		"Standard_NV4as_v4",
		"Standard_NV8as_v4",
		"Standard_NV16as_v4",
		"Standard_NV32as_v4",
		"Standard_L8s_v3",
		"Standard_L16s_v3",
		"Standard_L32s_v3",
		"Standard_L48s_v3",
		"Standard_L64s_v3",
		"Standard_L80s_v3",
		"Standard_E64i_v3",
		"Standard_E64is_v3",
		"Standard_G1",
		"Standard_G2",
		"Standard_G3",
		"Standard_G4",
		"Standard_G5",
		"Standard_GS1",
		"Standard_GS2",
		"Standard_GS3",
		"Standard_GS4",
		"Standard_GS4-4",
		"Standard_GS4-8",
		"Standard_GS5",
		"Standard_GS5-8",
		"Standard_GS5-16",
		"Standard_L4s",
		"Standard_L8s",
		"Standard_L16s",
		"Standard_L32s",
		"Standard_DC2as_v5",
		"Standard_DC4as_v5",
		"Standard_DC8as_v5",
		"Standard_DC16as_v5",
		"Standard_DC32as_v5",
		"Standard_DC48as_v5",
		"Standard_DC64as_v5",
		"Standard_DC96as_v5",
		"Standard_DC2ads_v5",
		"Standard_DC4ads_v5",
		"Standard_DC8ads_v5",
		"Standard_DC16ads_v5",
		"Standard_DC32ads_v5",
		"Standard_DC48ads_v5",
		"Standard_DC64ads_v5",
		"Standard_DC96ads_v5",
		"Standard_EC2as_v5",
		"Standard_EC4as_v5",
		"Standard_EC8as_v5",
		"Standard_EC16as_v5",
		"Standard_EC20as_v5",
		"Standard_EC32as_v5",
		"Standard_EC48as_v5",
		"Standard_EC64as_v5",
		"Standard_EC96as_v5",
		"Standard_EC96ias_v5",
		"Standard_EC2ads_v5",
		"Standard_EC4ads_v5",
		"Standard_EC8ads_v5",
		"Standard_EC16ads_v5",
		"Standard_EC20ads_v5",
		"Standard_EC32ads_v5",
		"Standard_EC48ads_v5",
		"Standard_EC64ads_v5",
		"Standard_EC96ads_v5",
		"Standard_EC96iads_v5",
		"Standard_E96ias_v4",
		"Standard_NC6s_v3",
		"Standard_NC12s_v3",
		"Standard_NC24rs_v3",
		"Standard_NC24s_v3",
		"Standard_NV6s_v2",
		"Standard_NV12s_v2",
		"Standard_NV24s_v2",
		"Standard_NV12s_v3",
		"Standard_NV24s_v3",
		"Standard_NV48s_v3",
		"Standard_M208ms_v2",
		"Standard_M208s_v2",
		"Standard_M416-208s_v2",
		"Standard_M416s_v2",
		"Standard_M416-208ms_v2",
		"Standard_M416ms_v2",
		"Standard_D2plds_v5",
		"Standard_D4plds_v5",
		"Standard_D8plds_v5",
		"Standard_D16plds_v5",
		"Standard_D32plds_v5",
		"Standard_D48plds_v5",
		"Standard_D64plds_v5",
		"Standard_D2pls_v5",
		"Standard_D4pls_v5",
		"Standard_D8pls_v5",
		"Standard_D16pls_v5",
		"Standard_D32pls_v5",
		"Standard_D48pls_v5",
		"Standard_D64pls_v5",
		"Standard_D2pds_v5",
		"Standard_D4pds_v5",
		"Standard_D8pds_v5",
		"Standard_D16pds_v5",
		"Standard_D32pds_v5",
		"Standard_D48pds_v5",
		"Standard_D64pds_v5",
		"Standard_D2ps_v5",
		"Standard_D4ps_v5",
		"Standard_D8ps_v5",
		"Standard_D16ps_v5",
		"Standard_D32ps_v5",
		"Standard_D48ps_v5",
		"Standard_D64ps_v5",
		"Standard_E2pds_v5",
		"Standard_E4pds_v5",
		"Standard_E8pds_v5",
		"Standard_E16pds_v5",
		"Standard_E20pds_v5",
		"Standard_E32pds_v5",
		"Standard_E2ps_v5",
		"Standard_E4ps_v5",
		"Standard_E8ps_v5",
		"Standard_E16ps_v5",
		"Standard_E20ps_v5",
		"Standard_E32ps_v5",
		"Standard_DC1s_v3",
		"Standard_DC2s_v3",
		"Standard_DC4s_v3",
		"Standard_DC8s_v3",
		"Standard_DC16s_v3",
		"Standard_DC24s_v3",
		"Standard_DC32s_v3",
		"Standard_DC48s_v3",
		"Standard_DC1ds_v3",
		"Standard_DC2ds_v3",
		"Standard_DC4ds_v3",
		"Standard_DC8ds_v3",
		"Standard_DC16ds_v3",
		"Standard_DC24ds_v3",
		"Standard_DC32ds_v3",
		"Standard_DC48ds_v3",
		"Standard_NC24ads_A100_v4",
		"Standard_NC48ads_A100_v4",
		"Standard_NC96ads_A100_v4",
	}
	for _, validRegion := range validDisks {
		if validRegion == disk {
			return true
		}
	}
	return false
}

func isValidRegion(region string) bool {
	validRegions := []string{"eastus",
		"eastus2",
		"southcentralus",
		"westus2",
		"westus3",
		"australiaeast",
		"southeastasia",
		"northeurope",
		"swedencentral",
		"uksouth",
		"westeurope",
		"centralus",
		"southafricanorth",
		"centralindia",
		"eastasia",
		"japaneast",
		"koreacentral",
		"canadacentral",
		"francecentral",
		"germanywestcentral",
		"norwayeast",
		"switzerlandnorth",
		"uaenorth",
		"brazilsouth",
		"centraluseuap",
		"eastus2euap",
		"qatarcentral",
		"centralusstage",
		"eastusstage",
		"eastus2stage",
		"northcentralusstage",
		"southcentralusstage",
		"westusstage",
		"westus2stage",
		"asia",
		"asiapacific",
		"australia",
		"brazil",
		"canada",
		"europe",
		"france",
		"germany",
		"global",
		"india",
		"japan",
		"korea",
		"norway",
		"singapore",
		"southafrica",
		"switzerland",
		"uae",
		"uk",
		"unitedstates",
		"unitedstateseuap",
		"eastasiastage",
		"southeastasiastage",
		"brazilus",
		"eastusstg",
		"northcentralus",
		"westus",
		"jioindiawest",
		"devfabric",
		"westcentralus",
		"southafricawest",
		"australiacentral",
		"australiacentral2",
		"australiasoutheast",
		"japanwest",
		"jioindiacentral",
		"koreasouth",
		"southindia",
		"westindia",
		"canadaeast",
		"francesouth",
		"germanynorth",
		"norwaywest",
		"switzerlandwest",
		"ukwest",
		"uaecentral",
		"brazilsoutheast"}

	for _, validRegion := range validRegions {
		if validRegion == region {
			return true
		}
	}
	return false
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

func (obj *AzureProvider) CreateResourceGroup(ctx context.Context, logging log.Logger) (*armresources.ResourceGroupsClientCreateOrUpdateResponse, error) {
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
	logging.Info("Created resource group", *resourceGroup.Name)
	return &resourceGroup, nil
}

func (obj *AzureProvider) DeleteResourceGroup(ctx context.Context, logging log.Logger) error {
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

	logging.Info("Deleted resource group", obj.Config.ResourceGroupName)
	return nil
}

func (obj *AzureProvider) CreateSubnet(ctx context.Context, logging log.Logger, subnetName string) (*armnetwork.Subnet, error) {
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
	logging.Info("Created subnet", *resp.Name)
	return &resp.Subnet, nil
}

func (obj *AzureProvider) DeleteSubnet(ctx context.Context, logging log.Logger, subnetName string) error {
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

	logging.Info("Deleted subnet", subnetName)
	return nil
}

func (obj *AzureProvider) CreatePublicIP(ctx context.Context, logging log.Logger, publicIPName string) (*armnetwork.PublicIPAddress, error) {
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
	logging.Info("Created public IP address", *resp.Name)
	return &resp.PublicIPAddress, err
}

func (obj *AzureProvider) DeleteAllPublicIP(ctx context.Context, logging log.Logger) error {
	for _, interfaceName := range obj.Config.InfoControlPlanes.PublicIPNames {
		if err := obj.DeletePublicIP(ctx, logging, interfaceName); err != nil {
			return err
		}
	}
	for _, interfaceName := range obj.Config.InfoWorkerPlanes.PublicIPNames {
		if err := obj.DeletePublicIP(ctx, logging, interfaceName); err != nil {
			return err
		}
	}

	if len(obj.Config.InfoDatabase.PublicIPName) != 0 {
		if err := obj.DeletePublicIP(ctx, logging, obj.Config.InfoDatabase.PublicIPName); err != nil {
			return err
		}
	}
	if len(obj.Config.InfoLoadBalancer.PublicIPName) != 0 {
		if err := obj.DeletePublicIP(ctx, logging, obj.Config.InfoLoadBalancer.PublicIPName); err != nil {
			return err
		}
	}
	logging.Info("Deleted all Public IPs", "")
	return nil
}

func (obj *AzureProvider) DeletePublicIP(ctx context.Context, logging log.Logger, publicIPName string) error {
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

	logging.Info("Deleted the pubIP", publicIPName)
	return nil
}

func (obj *AzureProvider) DeleteSSHKeyPair(ctx context.Context, logging log.Logger) error {
	sshClient, err := armcompute.NewSSHPublicKeysClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}
	_, err = sshClient.Delete(ctx, obj.Config.ResourceGroupName, obj.Config.SSHKeyName, nil)
	if err != nil {
		return err
	}

	logging.Info("Deleted the ssh", obj.Config.SSHKeyName)
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

func (obj *AzureProvider) CreateNetworkInterface(ctx context.Context, logging log.Logger, resourceName, nicName string, subnetID string, publicIPID string, networkSecurityGroupID string) (*armnetwork.Interface, error) {
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
	logging.Info("Created network interface", *resp.Name)
	return &resp.Interface, err
}

func (obj *AzureProvider) DeleteAllNetworkInterface(ctx context.Context, logging log.Logger) error {
	for _, interfaceName := range obj.Config.InfoControlPlanes.NetworkInterfaceNames {
		if err := obj.DeleteNetworkInterface(ctx, logging, interfaceName); err != nil {
			return err
		}
	}
	for _, interfaceName := range obj.Config.InfoWorkerPlanes.NetworkInterfaceNames {
		if err := obj.DeleteNetworkInterface(ctx, logging, interfaceName); err != nil {
			return err
		}
	}

	if len(obj.Config.InfoDatabase.NetworkInterfaceName) != 0 {
		if err := obj.DeleteNetworkInterface(ctx, logging, obj.Config.InfoDatabase.NetworkInterfaceName); err != nil {
			return err
		}
	}
	if len(obj.Config.InfoLoadBalancer.NetworkInterfaceName) != 0 {
		if err := obj.DeleteNetworkInterface(ctx, logging, obj.Config.InfoLoadBalancer.NetworkInterfaceName); err != nil {
			return err
		}
	}
	logging.Info("Deleted all network interfaces", "")
	return nil
}

func (obj *AzureProvider) DeleteNetworkInterface(ctx context.Context, logging log.Logger, nicName string) error {
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
	logging.Info("Deleted the nic", nicName)

	return nil
}

func (obj *AzureProvider) DeleteAllNSG(ctx context.Context, logging log.Logger) error {
	if len(obj.Config.InfoControlPlanes.NetworkSecurityGroupName) != 0 {
		if err := obj.DeleteNSG(ctx, logging, obj.Config.InfoControlPlanes.NetworkSecurityGroupName); err != nil {
			return err
		}
	}
	if len(obj.Config.InfoWorkerPlanes.NetworkSecurityGroupName) != 0 {
		if err := obj.DeleteNSG(ctx, logging, obj.Config.InfoWorkerPlanes.NetworkSecurityGroupName); err != nil {
			return err
		}
	}

	if len(obj.Config.InfoDatabase.NetworkSecurityGroupName) != 0 {
		if err := obj.DeleteNSG(ctx, logging, obj.Config.InfoDatabase.NetworkSecurityGroupName); err != nil {
			return err
		}
	}
	if len(obj.Config.InfoLoadBalancer.NetworkSecurityGroupName) != 0 {
		if err := obj.DeleteNSG(ctx, logging, obj.Config.InfoLoadBalancer.NetworkSecurityGroupName); err != nil {
			return err
		}
	}
	logging.Info("Deleted all network security groups", "")
	return nil
}

func (obj *AzureProvider) DeleteNSG(ctx context.Context, logging log.Logger, nsgName string) error {
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
	logging.Info("Deleted the nsg", nsgName)
	return nil
}

func (obj *AzureProvider) CreateNSG(ctx context.Context, logging log.Logger, nsgName string, securityRules []*armnetwork.SecurityRule) (*armnetwork.SecurityGroup, error) {
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
	logging.Info("Created network security group", *resp.Name)
	return &resp.SecurityGroup, nil
}

func (obj *AzureProvider) DeleteVirtualNetwork(ctx context.Context, logging log.Logger) error {
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
	logging.Info("Deleted virtual network", obj.Config.VirtualNetworkName)
	return nil
}

func (obj *AzureProvider) CreateVirtualNetwork(ctx context.Context, logging log.Logger, virtualNetworkName string) (*armnetwork.VirtualNetwork, error) {
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
	logging.Info("Created virtual network", *resp.Name)
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

func (obj *AzureProvider) CreateVM(ctx context.Context, logging log.Logger, vmName, networkInterfaceID, diskName, script string) (*armcompute.VirtualMachine, error) {
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
	logging.Info("Created network virtual machine", *resp.Name)
	return &resp.VirtualMachine, nil
}

func (obj *AzureProvider) DeleteAllVMs(ctx context.Context, logging log.Logger) error {
	for _, instanceName := range obj.Config.InfoControlPlanes.Names {
		if err := obj.DeleteVM(ctx, logging, instanceName); err != nil {
			return err
		}
	}
	for _, instanceName := range obj.Config.InfoWorkerPlanes.Names {
		if err := obj.DeleteVM(ctx, logging, instanceName); err != nil {
			return err
		}
	}

	if len(obj.Config.InfoDatabase.Name) != 0 {
		if err := obj.DeleteVM(ctx, logging, obj.Config.InfoDatabase.Name); err != nil {
			return err
		}
	}
	if len(obj.Config.InfoLoadBalancer.Name) != 0 {
		if err := obj.DeleteVM(ctx, logging, obj.Config.InfoLoadBalancer.Name); err != nil {
			return err
		}
	}
	logging.Info("Deleted all virtual machines", "")
	return nil
}

func (obj *AzureProvider) DeleteVM(ctx context.Context, logging log.Logger, vmName string) error {
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

	logging.Info("Deleted the vm", vmName)
	return nil
}

func (obj *AzureProvider) DeleteAllDisks(ctx context.Context, logging log.Logger) error {
	for _, diskName := range obj.Config.InfoControlPlanes.DiskNames {
		if err := obj.DeleteDisk(ctx, logging, diskName); err != nil {
			return err
		}
	}
	for _, diskName := range obj.Config.InfoWorkerPlanes.DiskNames {
		if err := obj.DeleteDisk(ctx, logging, diskName); err != nil {
			return err
		}
	}

	if len(obj.Config.InfoDatabase.DiskName) != 0 {
		if err := obj.DeleteDisk(ctx, logging, obj.Config.InfoDatabase.DiskName); err != nil {
			return err
		}
	}
	if len(obj.Config.InfoLoadBalancer.DiskName) != 0 {
		if err := obj.DeleteDisk(ctx, logging, obj.Config.InfoLoadBalancer.DiskName); err != nil {
			return err
		}
	}
	logging.Info("Deleted all disks", "")
	return nil
}

func (obj *AzureProvider) DeleteDisk(ctx context.Context, logging log.Logger, diskName string) error {
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
	logging.Info("Deleted disk", diskName)
	return nil
}

// SaveKubeconfig stores the kubeconfig to state management file
func (obj *AzureProvider) SaveKubeconfig(logging log.Logger, kubeconfig string) error {
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
	logging.Info("💾 Kubeconfig", "")
	return nil
}
