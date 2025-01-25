// Copyright 2024 Ksctl Authors
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

package logger

import (
	"context"
	"github.com/ksctl/ksctl/pkg/consts"
)

type Logger interface {
	Print(ctx context.Context, msg string, v ...any)

	Success(ctx context.Context, msg string, v ...any)

	Note(ctx context.Context, msg string, v ...any)

	Warn(ctx context.Context, msg string, v ...any)

	Error(msg string, v ...any)

	Debug(ctx context.Context, msg string, v ...any)

	ExternalLogHandler(ctx context.Context, msgType CustomExternalLogLevel, message string)
	ExternalLogHandlerf(ctx context.Context, msgType CustomExternalLogLevel, format string, args ...interface{})

	NewError(ctx context.Context, msg string, v ...any) error

	Table(ctx context.Context, operation LogClusterDetail, data []ClusterDataForLogging)

	Box(ctx context.Context, title string, lines string)
}

type VMData struct {
	VMID         string
	VMName       string
	VMSize       string
	FirewallID   string
	FirewallName string
	SubnetID     string
	SubnetName   string
	PublicIP     string
	PrivateIP    string
}

type ClusterDataForLogging struct {
	Name            string
	CloudProvider   consts.KsctlCloud
	ClusterType     consts.KsctlClusterType
	K8sDistro       consts.KsctlKubernetes
	SSHKeyName      string
	SSHKeyID        string
	Region          string
	ResourceGrpName string
	NetworkName     string
	NetworkID       string
	ManagedK8sID    string
	ManagedK8sName  string
	WP              []VMData
	CP              []VMData
	DS              []VMData
	LB              VMData
	Mgt             VMData
	NoWP            int
	NoCP            int
	NoDS            int
	NoMgt           int
	K8sVersion      string
	EtcdVersion     string
	HAProxyVersion  string
	Apps            []string
	Cni             string
}
