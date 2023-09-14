package ksctlhttpserver

import (
	"context"
	"log"
	"os"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers"
	"github.com/kubesimplify/ksctl/api/utils"
	httpserver "github.com/kubesimplify/ksctl/httpserver/gen/httpserver"

	control_pkg "github.com/kubesimplify/ksctl/api/controllers"
)

var (
	cli        *resources.CobraCmd
	controller controllers.Controller
	dir        = "/app/ksctl-data"
)

// httpserver service example implementation.
// The example methods log the requests and return zero values.
type httpserversrvc struct {
	logger *log.Logger
}

// NewHttpserver returns the httpserver service implementation.
func NewHttpserver(logger *log.Logger) httpserver.Service {
	return &httpserversrvc{logger}
}

// CreateHa implements create ha.
func (s *httpserversrvc) CreateHa(ctx context.Context, p *httpserver.Metadata) (res *httpserver.Response, err error) {

	// TODO: use a different const for this rather than using test
	_ = os.Setenv(utils.KSCTL_TEST_DIR_ENABLED, dir)

	if err := os.MkdirAll(utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_CIVO, utils.CLUSTER_TYPE_HA), 0755); err != nil {
		return res, err
	}

	if err := os.MkdirAll(utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AZURE, utils.CLUSTER_TYPE_HA), 0755); err != nil {
		return res, err
	}

	// NOTE: move the above things to the pod level or docker level

	cli = &resources.CobraCmd{}
	controller = control_pkg.GenKsctlController()

	cli.Client.Metadata.ClusterName = p.ClusterName
	cli.Client.Metadata.StateLocation = utils.STORE_LOCAL
	cli.Client.Metadata.K8sDistro = p.Distro

	if _, err := control_pkg.InitializeStorageFactory(&cli.Client, true); err != nil {
		panic(err)
	}

	cli.Client.Metadata.Region = p.Region
	cli.Client.Metadata.Provider = p.Cloud

	cli.Client.Metadata.NoCP = int(*p.NoCp)
	cli.Client.Metadata.NoDS = int(*p.NoDs)
	cli.Client.Metadata.NoWP = int(*p.NoWp)
	cli.Client.Metadata.IsHA = true

	cli.Client.Metadata.ControlPlaneNodeType = *p.VMSizeCp
	cli.Client.Metadata.WorkerPlaneNodeType = *p.VMSizeWp
	cli.Client.Metadata.LoadBalancerNodeType = *p.VMSizeLb
	cli.Client.Metadata.DataStoreNodeType = *p.VMSizeDs

	// Return
	ok := true

	msg, errStr := controller.CreateHACluster(&cli.Client)
	if errStr != nil {
		ok = false
		err = errStr
	}

	resErr := errStr.Error()

	res = &httpserver.Response{OK: &ok, Errors: &resErr, Response: msg}
	s.logger.Print(msg)

	return
}

// DeleteHa implements delete ha.
func (s *httpserversrvc) DeleteHa(ctx context.Context, p *httpserver.Metadata) (res *httpserver.Response, err error) {

	_ = os.Setenv(utils.KSCTL_TEST_DIR_ENABLED, dir)

	if err := os.MkdirAll(utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_CIVO, utils.CLUSTER_TYPE_HA), 0755); err != nil {
		return res, err
	}

	if err := os.MkdirAll(utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AZURE, utils.CLUSTER_TYPE_HA), 0755); err != nil {
		return res, err
	}

	cli = &resources.CobraCmd{}
	controller = control_pkg.GenKsctlController()

	cli.Client.Metadata.ClusterName = p.ClusterName
	cli.Client.Metadata.StateLocation = utils.STORE_LOCAL
	cli.Client.Metadata.K8sDistro = p.Distro

	if _, err := control_pkg.InitializeStorageFactory(&cli.Client, true); err != nil {
		panic(err)
	}

	cli.Client.Metadata.IsHA = true
	cli.Client.Metadata.Region = p.Region
	cli.Client.Metadata.Provider = p.Cloud

	// Return
	ok := true

	msg, errStr := controller.DeleteHACluster(&cli.Client)
	if errStr != nil {
		ok = false
	}

	resErr := errStr.Error()

	res = &httpserver.Response{OK: &ok, Errors: &resErr, Response: msg}
	s.logger.Print(msg)

	return
}

// Scaledown implements scaledown.
func (s *httpserversrvc) Scaledown(ctx context.Context, p *httpserver.Metadata) (res *httpserver.Response, err error) {
	res = &httpserver.Response{}
	s.logger.Print("httpserver.scaledown")
	return
}

// Scaleup implements scaleup.
func (s *httpserversrvc) Scaleup(ctx context.Context, p *httpserver.Metadata) (res *httpserver.Response, err error) {
	res = &httpserver.Response{}
	s.logger.Print("httpserver.scaleup")
	return
}

// GetHealth implements get health.
func (s *httpserversrvc) GetHealth(ctx context.Context) (res *httpserver.Health, err error) {
	res = &httpserver.Health{}
	s.logger.Print("httpserver.get health")
	return
}

// GetClusters implements get clusters.
func (s *httpserversrvc) GetClusters(ctx context.Context) (res *httpserver.Response, err error) {
	res = &httpserver.Response{}
	s.logger.Print("httpserver.get clusters")
	return
}
