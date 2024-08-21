package storage

import "time"

type StateConfigurationKubeadm struct {
	KubeadmVersion string `json:"kubeadm_version" bson:"kubeadm_version"`

	CertificateKey string `json:"certificate_key" bson:"certificate_key"`

	BootstrapToken string `json:"bootstrap_token" bson:"bootstrap_token"`

	BootstrapTokenExpireTimeUtc time.Time `json:"bootstrap_token_expire_time_utc" bson:"bootstrap_token_expire_time_utc"`

	DiscoveryTokenCACertHash string `json:"discorvery_token_ca_cert_hash" bson:"discorvery_token_ca_cert_hash"`
}
