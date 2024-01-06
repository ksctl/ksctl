package types

type BaseInfra struct {
	IsCompleted bool `json:"status" bson:"status"`

	SSHID      string `json:"ssh_id" bson:"ssh_id"`
	SSHUser    string `json:"ssh_usr" bson:"ssh_usr"`
	SSHKeyName string `json:"sshkey_name" bson:"sshkey_name"`

	KubernetesDistro string `json:"k8s_distro" bson:"k8s_distro"`
	KubernetesVer    string `json:"k8s_version" bson:"k8s_version"`
}
