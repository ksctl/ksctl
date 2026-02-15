// Copyright 2025 Ksctl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package meta

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/ksctl/ksctl/v2/pkg/addons"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	awsPkg "github.com/ksctl/ksctl/v2/pkg/provider/aws"
)

func WithDefaultEC2() Option {
	return func(o *options) error {
		o.ec2Type = []string{
			// x86_64 instance families
			"c5*", "t3*", "m5*", "r5*", "m7i*", "m7a*",
			// ARM64 (Graviton) instance families
			"t4g*", "m6g*", "m7g*", "m8g*", "c6g*", "c7g*", "r6g*", "r7g*",
		}
		return nil
	}
}

func (m *AwsMeta) listOfVms(region string, opts ...Option) (out []provider.InstanceRegionOutput, _ error) {

	options := options{}
	for _, o := range opts {
		if err := o(&options); err != nil {
			return nil, err
		}
	}

	var vmTypes ec2.DescribeInstanceTypesOutput
	// https://github.com/aws/aws-sdk-go-v2/blob/service/ec2/v1.198.1/service/ec2/api_op_DescribeInstanceTypes.go#L31
	input := &ec2.DescribeInstanceTypesInput{
		Filters: []ec2Types.Filter{
			{
				Name:   aws.String("current-generation"),
				Values: []string{"true"},
			},
			{
				Name:   aws.String("supported-root-device-type"),
				Values: []string{"ebs"},
			},
			{
				Name:   aws.String("supported-usage-class"),
				Values: []string{"on-demand"},
			},
			{
				Name:   aws.String("processor-info.supported-architecture"),
				Values: []string{"x86_64", "arm64"},
			},
		},
	}

	if len(options.ec2Type) > 0 {
		input.Filters = append(input.Filters, ec2Types.Filter{
			Name:   aws.String("instance-type"),
			Values: options.ec2Type,
		})
	}

	session, err := m.GetNewSession(region)
	if err != nil {
		return nil, err
	}

	ec2C := ec2.NewFromConfig(*session)

	for {
		output, err := ec2C.DescribeInstanceTypes(m.ctx, input)
		if err != nil {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidCloudVMSize,
				m.l.NewError(m.ctx, "Error in fetching instance types", "Reason", err),
			)
		}
		vmTypes.InstanceTypes = append(vmTypes.InstanceTypes, output.InstanceTypes...)
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	analyseVMType := func(vmTypeSku string) provider.MachineCategory {

		var category provider.MachineCategory
		// Compute Intensive: c5 (x86), c6g, c7g (ARM64)
		if strings.HasPrefix(vmTypeSku, "c5") ||
			strings.HasPrefix(vmTypeSku, "c6g") ||
			strings.HasPrefix(vmTypeSku, "c7g") {
			category = provider.ComputeIntensive
		} else if strings.HasPrefix(vmTypeSku, "t3") ||
			strings.HasPrefix(vmTypeSku, "t4g") {
			// Burst: t3 (x86), t4g (ARM64)
			category = provider.Burst
		} else if strings.HasPrefix(vmTypeSku, "m") {
			// General Purpose: m5, m7i, m7a (x86), m6g, m7g, m8g (ARM64)
			category = provider.GeneralPurpose
		} else if strings.HasPrefix(vmTypeSku, "r5") ||
			strings.HasPrefix(vmTypeSku, "r6g") ||
			strings.HasPrefix(vmTypeSku, "r7g") {
			// Memory Intensive: r5 (x86), r6g, r7g (ARM64)
			category = provider.MemoryIntensive
		} else {
			return provider.Unknown
		}

		return category
	}

	for _, vm := range vmTypes.InstanceTypes {
		var arch provider.MachineArchitecture
		for _, arc := range vm.ProcessorInfo.SupportedArchitectures {
			if arc == ec2Types.ArchitectureTypeX8664 {
				arch = provider.ArchAmd64
				break
			}
			if arc == ec2Types.ArchitectureTypeArm64 {
				arch = provider.ArchArm64
				break
			}
		}

		category := analyseVMType(string(vm.InstanceType))

		out = append(
			out,
			provider.InstanceRegionOutput{
				Sku:                    string(vm.InstanceType),
				Description:            string(vm.InstanceType),
				Category:               category,
				VCpus:                  *vm.VCpuInfo.DefaultVCpus,
				Memory:                 int32(*vm.MemoryInfo.SizeInMiB / 1024),
				CpuArch:                arch,
				EphemeralOSDiskSupport: *vm.InstanceStorageSupported,
			},
		)
	}

	return out, nil
}

func (m *AwsMeta) listManagedOfferingsK8sVersions(regionSku string) ([]string, error) {
	input := &eks.DescribeAddonVersionsInput{
		AddonName:         aws.String("vpc-cni"),
		KubernetesVersion: aws.String(""),
	}

	session, err := m.GetNewSession(regionSku)
	if err != nil {
		return nil, err
	}

	resp, err := eks.NewFromConfig(*session).DescribeAddonVersions(m.ctx, input)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			m.l.NewError(m.ctx, "Error Describing Addon Versions", "Reason", err),
		)
	}

	versions := make(map[string]struct{})
	for _, addon := range resp.Addons {
		for _, addonVersion := range addon.AddonVersions {
			for _, k8sVersion := range addonVersion.Compatibilities {
				if k8sVersion.ClusterVersion != nil {
					versions[*k8sVersion.ClusterVersion] = struct{}{}
				}
			}
		}
	}

	var s []string
	for k := range versions {
		s = append(s, k)
	}

	return s, nil
}

func (m *AwsMeta) listManagedOfferingsCNIPlugins(_ string) (addons.ClusterAddons, string, error) {
	v, d := awsPkg.GetManagedCNIAddons()
	return v, d, nil
}
