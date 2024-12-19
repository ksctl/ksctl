package statefile

type StateConfigurationLocal struct {
	B               BaseInfra `json:"b" bson:"b"`
	Nodes           int       `json:"nodes" bson:"nodes"`
	ManagedNodeSize string    `json:"managed_node_size" bson:"managed_node_size"`
}
