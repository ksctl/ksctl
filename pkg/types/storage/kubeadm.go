package storage

import "time"

type StateConfigurationKubeadm struct {
	// CertificateKey is only valid till 24 hours
	CertificateKey string `json:"certificate_key" bson:"certificate_key"`

	// BootstrapToken cryptographic random string
	BootstrapToken string `json:"bootstrap_token" bson:"bootstrap_token"`

	BootstrapTokenCreationTimeUtc time.Time `json:"bootstrap_token_creation_time_utc" bson:"bootstrap_token_creation_time_utc"`

	DiscoveryTokenCACertHash string `json:"discorvery_token_ca_cert_hash" bson:"discorvery_token_ca_cert_hash"`
}
