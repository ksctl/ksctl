package types

type StateConfigurationK3s struct {
	B BaseK8sBootstrap `json:"b" bson:"b"`

	K3sToken          string `json:"k3s_token" bson:"k3s_token"`
	DataStoreEndPoint string `json:"datastore_endpoint" bson:"datastore_endpoint"` // DEPRECATED !! for etcd as datastore
}
