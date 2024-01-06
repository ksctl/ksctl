package types

type CivoStateVMs struct {
	VMIDs      []string `json:"vm_ids" bson:"vm_ids"`
	PrivateIPs []string `json:"private_ips" bson:"private_ips"`
	PublicIPs  []string `json:"public_ips" bson:"public_ips"`
	Hostnames  []string `json:"hostnames" bson:"hostnames"`
}

type CivoStateVM struct {
	VMID      string `json:"vm_id" bson:"vm_id"`
	PrivateIP string `json:"private_ip" bson:"private_ip"`
	PublicIP  string `json:"public_ip" bson:"public_ip"`
	HostName  string `json:"hostname" bson:"hostname"`
}

type CredentialsCivo struct {
	Token string `json:"token" bson:"token"`
}

type StateConfigurationCivo struct {
	B BaseInfra `json:"b" bson:"b"`

	ManagedClusterID string `json:"managed_cluster_id" bson:"managed_cluster_id"`
	NoManagedNodes   int    `json:"no_managed_cluster_nodes" bson:"no_managed_cluster_nodes"`

	FirewallIDControlPlanes string `json:"fwidcontrolplanenode" bson:"fwidcontrolplanenode"`
	FirewallIDWorkerNodes   string `json:"fwidworkernode" bson:"fwidworkernode"`
	FirewallIDLoadBalancer  string `json:"fwidloadbalancenode" bson:"fwidloadbalancenode"`
	FirewallIDDatabaseNodes string `json:"fwiddatabasenode" bson:"fwiddatabasenode"`
	NetworkID               string `json:"clusternetworkid" bson:"clusternetworkid"`

	InfoControlPlanes CivoStateVMs `json:"info_control_planes" bson:"info_control_planes"`
	InfoWorkerPlanes  CivoStateVMs `json:"info_worker_planes" bson:"info_worker_planes"`
	InfoDatabase      CivoStateVMs `json:"info_database" bson:"info_database"`
	InfoLoadBalancer  CivoStateVM  `json:"info_load_balancer" bson:"info_load_balancer"`
}
