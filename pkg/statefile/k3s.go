package statefile

type StateConfigurationK3s struct {
	K3sToken string `json:"k3s_token" bson:"k3s_token"`

	K3sVersion string `json:"k3s_version" bson:"k3s_version"`
}
