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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/provider"
)

func (m *AwsMeta) listOfVms(region string) (out []provider.InstanceRegionOutput, _ error) {
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

		out = append(
			out,
			provider.InstanceRegionOutput{
				Sku:         string(vm.InstanceType),
				Description: string(vm.InstanceType),
				VCpus:       *vm.VCpuInfo.DefaultVCpus,
				//Memory:      int32(*vm.MemoryInfo.SizeInMiB / 1024),
				Memory:                 int32(*vm.MemoryInfo.SizeInMiB), // TODO: please check what is the value
				CpuArch:                arch,
				EphemeralOSDiskSupport: *vm.InstanceStorageSupported,
			},
		)
	}

	return out, nil
}
