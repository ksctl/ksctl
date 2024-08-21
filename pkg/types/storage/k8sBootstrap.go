package storage

import "github.com/ksctl/ksctl/pkg/types/controllers/cloud"

type BaseK8sBootstrap struct {
	SSHInfo    cloud.SSHInfo `json:"cloud_ssh_info" bson:"cloud_ssh_info"`
	PublicIPs  Instances     `json:"cloud_public_ips" bson:"cloud_public_ips"`
	PrivateIPs Instances     `json:"cloud_private_ips" bson:"cloud_private_ips"`

	HAProxyVersion string `json:"haproxy_version" bson:"haproxy_version"`
	EtcdVersion    string `json:"etcd_version" bson:"etcd_version"`

	CACert   string `json:"ca_cert" bson:"ca_cert"`
	EtcdCert string `json:"etcd_cert" bson:"etcd_cert"`
	EtcdKey  string `json:"etcd_key" bson:"etcd_key"`
}
