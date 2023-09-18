package aws

import "github.com/aws/aws-sdk-go-v2/service/ec2"

func ProvideClient() AwsGo {
	return &AwsGoClient{}
}

/* TODO figure out about pull until done funtions */

type AwsGo interface {
	//SetRegion(string)

	ListLocations() ([]string, error)

	ListVMTypes() ([]string, error)

	BeginCreateVpc() (*ec2.CreateVpcOutput, error)

	BeginCreateVirtNet() error

	BeginDeleteVirtNet() error

	BeginCreateSubNet() error

	BeginDeleteSubNet() error

	CreateSSHKey() error

	DeleteSSHKey() error

	BeginCreateVM() error

	BeginDeleteVM() error

	BeginCreatePubIP() error

	BeginDeletePubIP() error

	BeginCreateNIC() error

	BeginDeleteNIC() error

	BeginDeleteSecurityGrp() error
	BeginCreateSecurityGrp() error
	SetRegion(role string)
	SetResourceGrp(name string)
}

type AwsGoClient struct {
	ACESSKEYID     string
	ACESSKEYSECRET string
	Region         string
	Vpc            string
}

// BeginCreateNIC implements AwsGo.
func (awsclient *AwsGoClient) BeginCreateNIC() error {
	panic("unimplemented")
}

// BeginCreatePubIP implements AwsGo.
func (AwsGoClient) BeginCreatePubIP() error {
	panic("unimplemented")
}

// BeginCreateSecurityGrp implements AwsGo.
func (AwsGoClient) BeginCreateSecurityGrp() error {
	panic("unimplemented")
}

// BeginCreateSubNet implements AwsGo.
func (AwsGoClient) BeginCreateSubNet() error {
	panic("unimplemented")
}

// BeginCreateVM implements AwsGo.
func (AwsGoClient) BeginCreateVM() error {
	panic("unimplemented")
}

// BeginCreateVirtNet implements AwsGo.
func (AwsGoClient) BeginCreateVirtNet() error {
	panic("unimplemented")
}

// BeginCreateVpc implements AwsGo.
func (AwsGoClient) BeginCreateVpc() (*ec2.CreateVpcOutput, error) {
	panic("unimplemented")
}

// BeginDeleteNIC implements AwsGo.
func (AwsGoClient) BeginDeleteNIC() error {
	panic("unimplemented")
}

// BeginDeletePubIP implements AwsGo.
func (AwsGoClient) BeginDeletePubIP() error {
	panic("unimplemented")
}

// BeginDeleteSecurityGrp implements AwsGo.
func (AwsGoClient) BeginDeleteSecurityGrp() error {
	panic("unimplemented")
}

// BeginDeleteSubNet implements AwsGo.
func (AwsGoClient) BeginDeleteSubNet() error {
	panic("unimplemented")
}

// BeginDeleteVM implements AwsGo.
func (AwsGoClient) BeginDeleteVM() error {
	panic("unimplemented")
}

// BeginDeleteVirtNet implements AwsGo.
func (AwsGoClient) BeginDeleteVirtNet() error {
	panic("unimplemented")
}

// CreateSSHKey implements AwsGo.
func (AwsGoClient) CreateSSHKey() error {
	panic("unimplemented")
}

// DeleteSSHKey implements AwsGo.
func (AwsGoClient) DeleteSSHKey() error {
	panic("unimplemented")
}

// ListLocations implements AwsGo.
func (AwsGoClient) ListLocations() ([]string, error) {
	panic("unimplemented")
}

// ListVMTypes implements AwsGo.
func (AwsGoClient) ListVMTypes() ([]string, error) {
	panic("unimplemented")
}

// SetRegion implements AwsGo.
func (AwsGoClient) SetRegion(string) {
	panic("unimplemented")
}

// SetResourceGrp implements AwsGo.
func (AwsGoClient) SetResourceGrp(string) {
	panic("unimplemented")
}
