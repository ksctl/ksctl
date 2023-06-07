package providers

type CivoProvider struct {
	ClusterName string `json:"cluster_name"`
	APIKey      string `json:"api_key"`
	HACluster   bool   `json:"ha_cluster"`
	Region      string `json:"region"`
	//Spec        util.Machine `json:"spec"`
	Application string `json:"application"`
	CNIPlugin   string `json:"cni_plugin"`
}
