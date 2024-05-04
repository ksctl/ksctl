package types

type AzureStateVMs struct {
	Names                    []string `json:"names" bson:"names"`
	NetworkSecurityGroupName string   `json:"network_security_group_name" bson:"network_security_group_name"`
	NetworkSecurityGroupID   string   `json:"network_security_group_id" bson:"network_security_group_id"`
	DiskNames                []string `json:"disk_names" bson:"disk_names"`
	PublicIPNames            []string `json:"public_ip_names" bson:"public_ip_names"`
	PublicIPIDs              []string `json:"public_ip_ids" bson:"public_ip_ids"`
	PrivateIPs               []string `json:"private_ips" bson:"private_ips"`
	PublicIPs                []string `json:"public_ips" bson:"public_ips"`
	NetworkInterfaceNames    []string `json:"network_interface_names" bson:"network_interface_names"`
	NetworkInterfaceIDs      []string `json:"network_interface_ids" bson:"network_interface_ids"`
	Hostnames                []string `json:"hostnames" bson:"hostnames"`
}

type AzureStateVM struct {
	Name                     string `json:"name" bson:"name"`
	NetworkSecurityGroupName string `json:"network_security_group_name" bson:"network_security_group_name"`
	NetworkSecurityGroupID   string `json:"network_security_group_id" bson:"network_security_group_id"`
	DiskName                 string `json:"disk_name" bson:"disk_name"`
	PublicIPName             string `json:"public_ip_name" bson:"public_ip_name"`
	PublicIPID               string `json:"public_ip_id" bson:"public_ip_id"`
	NetworkInterfaceName     string `json:"network_interface_name" bson:"network_interface_name"`
	NetworkInterfaceID       string `json:"network_interface_id" bson:"network_interface_id"`
	PrivateIP                string `json:"private_ip" bson:"private_ip"`
	PublicIP                 string `json:"public_ip" bson:"public_ip"`
	HostName                 string `json:"hostname" bson:"hostname"`
}

type CredentialsAzure struct {
	SubscriptionID string `json:"subscription_id" bson:"subscription_id"`
	TenantID       string `json:"tenant_id" bson:"tenant_id"`
	ClientID       string `json:"client_id" bson:"client_id"`
	ClientSecret   string `json:"client_secret" bson:"client_secret"`
}

type StateConfigurationAzure struct {
	B BaseInfra `json:"b" bson:"b"`

	ResourceGroupName string `json:"resource_group_name" bson:"resource_group_name"`

	ManagedClusterName string `json:"managed_cluster_name" bson:"managed_cluster_name"`
	NoManagedNodes     int    `json:"no_managed_cluster_nodes" bson:"no_managed_cluster_nodes"`

	SubnetName         string `json:"subnet_name" bson:"subnet_name"`
	SubnetID           string `json:"subnet_id" bson:"subnet_id"`
	VirtualNetworkName string `json:"virtual_network_name" bson:"virtual_network_name"`
	VirtualNetworkID   string `json:"virtual_network_id" bson:"virtual_network_id"`
	NetCidr            string `json:"net_cidr" bson:"net_cidr"`

	InfoControlPlanes AzureStateVMs `json:"info_control_planes" bson:"info_control_planes"`
	InfoWorkerPlanes  AzureStateVMs `json:"info_worker_planes" bson:"info_worker_planes"`
	InfoDatabase      AzureStateVMs `json:"info_database" bson:"info_database"`
	InfoLoadBalancer  AzureStateVM  `json:"info_load_balancer" bson:"info_load_balancer"`
}
