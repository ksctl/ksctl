package types

type StateConfigurationLocal struct {
	B     BaseInfra `json:"b" bson:"b"`
	Nodes int       `json:"nodes" bson:"nodes"`
}
