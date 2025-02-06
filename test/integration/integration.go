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

package integration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	controllerCommon "github.com/ksctl/ksctl/v2/pkg/handler/cluster/common"
	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"
	controllerManaged "github.com/ksctl/ksctl/v2/pkg/handler/cluster/managed"
	controllerSelfManaged "github.com/ksctl/ksctl/v2/pkg/handler/cluster/selfmanaged"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
)

var (
	cli *controller.Client
	dir = filepath.Join(os.TempDir(), "ksctl-black-box-test")
	ctx context.Context
)

func InitCore() (err error) {
	ctx = context.WithValue(
		context.Background(),
		consts.KsctlContextUserID,
		"demo",
	)
	ctx = context.WithValue(
		ctx,
		consts.KsctlCustomDirLoc,
		dir,
	)
	ctx = context.WithValue(
		ctx,
		consts.KsctlTestFlagKey,
		"true",
	)

	awsC, err := json.Marshal(statefile.CredentialsAws{
		AccessKeyId:     "fake",
		SecretAccessKey: "fake",
	})
	if err != nil {
		return err
	}

	azC, err := json.Marshal(statefile.CredentialsAzure{
		SubscriptionID: "fake",
		ClientID:       "fake",
		ClientSecret:   "fake",
		TenantID:       "fake",
	})
	if err != nil {
		return err
	}

	ctx = context.WithValue(
		ctx,
		consts.KsctlAwsCredentials,
		awsC,
	)
	ctx = context.WithValue(
		ctx,
		consts.KsctlAzureCredentials,
		azC,
	)

	cli = new(controller.Client)

	cli.Metadata.ClusterName = "fake"
	cli.Metadata.StateLocation = consts.StoreLocal
	cli.Metadata.K8sDistro = consts.K8sK3s

	return
}

func ExecuteKsctlSpecificRun() error {
	log := logger.NewStructuredLogger(-1, os.Stdout)

	controller, err := controllerCommon.NewController(
		ctx,
		log,
		cli,
	)
	if err != nil {
		return err
	}

	if _, err := controller.Switch(); err != nil {
		return err
	}

	if _, err := controller.InfoCluster(); err != nil {
		return err
	}

	if err := controller.GetCluster(); err != nil {
		return err
	}
	return nil
}

func ExecuteManagedRun() error {
	log := logger.NewStructuredLogger(-1, os.Stdout)

	controller, err := controllerManaged.NewController(
		ctx,
		log,
		cli,
	)
	if err != nil {
		return err
	}

	if err := controller.Create(); err != nil {
		return err
	}

	if err := ExecuteKsctlSpecificRun(); err != nil {
		return err
	}

	if err := controller.Delete(); err != nil {
		return err
	}
	return nil
}

func ExecuteHARun() error {
	log := logger.NewStructuredLogger(-1, os.Stdout)

	controller, err := controllerSelfManaged.NewController(
		ctx,
		log,
		cli,
	)
	if err != nil {
		return err
	}

	if err := controller.Create(); err != nil {
		return err
	}

	if err := ExecuteKsctlSpecificRun(); err != nil {
		return err
	}

	cli.Metadata.NoWP = 0
	if err := controller.DeleteWorkerNodes(); err != nil {
		return err
	}

	if err := ExecuteKsctlSpecificRun(); err != nil {
		return err
	}

	cli.Metadata.NoWP = 1
	if err := controller.AddWorkerNodes(); err != nil {
		return err
	}

	if err := ExecuteKsctlSpecificRun(); err != nil {
		return err
	}

	if err := controller.Delete(); err != nil {
		return err
	}
	return nil
}

func AwsTestingManaged() error {
	cli.Metadata.Region = "fake-region"
	cli.Metadata.Provider = consts.CloudAws
	cli.Metadata.ClusterType = consts.ClusterTypeMang
	cli.Metadata.ManagedNodeType = "fake"
	cli.Metadata.NoMP = 2
	cli.Metadata.K8sVersion = "1.30"

	return ExecuteManagedRun()
}

func AzureTestingManaged() error {
	cli.Metadata.Region = "fake"
	cli.Metadata.Provider = consts.CloudAzure
	cli.Metadata.ManagedNodeType = "fake"
	cli.Metadata.ClusterType = consts.ClusterTypeMang
	cli.Metadata.NoMP = 2
	cli.Metadata.K8sVersion = "1.27"

	return ExecuteManagedRun()
}

func LocalTestingManaged() error {
	cli.Metadata.Provider = consts.CloudLocal
	cli.Metadata.NoMP = 5
	cli.Metadata.K8sVersion = "1.30.0"
	cli.Metadata.ClusterType = consts.ClusterTypeMang

	return ExecuteManagedRun()
}

func AzureTestingHAKubeadm() error {
	cli.Metadata.LoadBalancerNodeType = "fake"
	cli.Metadata.ControlPlaneNodeType = "fake"
	cli.Metadata.WorkerPlaneNodeType = "fake"
	cli.Metadata.DataStoreNodeType = "fake"

	cli.Metadata.ClusterType = consts.ClusterTypeSelfMang

	cli.Metadata.Region = "fake"
	cli.Metadata.Provider = consts.CloudAzure
	cli.Metadata.K8sDistro = consts.K8sKubeadm
	cli.Metadata.NoCP = 3
	cli.Metadata.NoWP = 1
	cli.Metadata.NoDS = 3

	return ExecuteHARun()
}

func AzureTestingHAK3s() error {
	cli.Metadata.LoadBalancerNodeType = "fake"
	cli.Metadata.ControlPlaneNodeType = "fake"
	cli.Metadata.WorkerPlaneNodeType = "fake"
	cli.Metadata.DataStoreNodeType = "fake"

	cli.Metadata.ClusterType = consts.ClusterTypeSelfMang

	cli.Metadata.Region = "fake"
	cli.Metadata.Provider = consts.CloudAzure
	cli.Metadata.K8sDistro = consts.K8sK3s
	cli.Metadata.NoCP = 3
	cli.Metadata.NoWP = 1
	cli.Metadata.NoDS = 3

	return ExecuteHARun()
}

func AwsTestingHA() error {
	cli.Metadata.LoadBalancerNodeType = "fake"
	cli.Metadata.ControlPlaneNodeType = "fake"
	cli.Metadata.WorkerPlaneNodeType = "fake"
	cli.Metadata.DataStoreNodeType = "fake"

	cli.Metadata.ClusterType = consts.ClusterTypeSelfMang

	cli.Metadata.Region = "fake-region"
	cli.Metadata.Provider = consts.CloudAws
	cli.Metadata.NoCP = 3
	cli.Metadata.NoWP = 1
	cli.Metadata.NoDS = 3

	return ExecuteHARun()
}
