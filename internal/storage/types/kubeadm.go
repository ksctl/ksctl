package types

type StateConfigurationKubeadm struct {
	B BaseK8sBootstrap `json:"b" bson:"b"`

	// CertificateKey is only valid till 24 hours
	CertificateKey string `json:"certificate_key" bson:"certificate_key"`

	// BootstrapToken cryptographic random string
	BootstrapToken string `json:"bootstrap_token" bson:"bootstrap_token"`

	DiscoveryTokenCACertHash string `json:"discorvery_token_ca_cert_hash" bson:"discorvery_token_ca_cert_hash"`
}
