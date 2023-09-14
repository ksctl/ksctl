package ksctlhttpserver

import (
	"context"
	"log"

	control_pkg "github.com/kubesimplify/ksctl/api/controllers"
	azure_pkg "github.com/kubesimplify/ksctl/api/provider/azure"
	civo_pkg "github.com/kubesimplify/ksctl/api/provider/civo"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers"
	cloudController "github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
	"github.com/kubesimplify/ksctl/api/utils"
	httpserver "github.com/kubesimplify/ksctl/httpserver/gen/httpserver"
)

var (
	cli        *resources.CobraCmd
	controller controllers.Controller
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

	cli = &resources.CobraCmd{}
	controller = control_pkg.GenKsctlController()

	cli.Client.Metadata.ClusterName = p.ClusterName
	cli.Client.Metadata.StateLocation = utils.STORE_LOCAL
	cli.Client.Metadata.K8sDistro = p.Distro
	cli.Client.Metadata.K8sVersion = "1.27.1"

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
	errStr := ""

	msg, err := controller.CreateHACluster(&cli.Client)
	if err != nil {
		ok = false
		errStr = err.Error()
	}

	res = &httpserver.Response{OK: &ok, Errors: &errStr, Response: msg}
	s.logger.Print(msg)

	return
}

// DeleteHa implements delete ha.
func (s *httpserversrvc) DeleteHa(ctx context.Context, p *httpserver.Metadata) (res *httpserver.Response, err error) {

	cli = &resources.CobraCmd{}
	controller = control_pkg.GenKsctlController()

	cli.Client.Metadata.ClusterName = p.ClusterName
	cli.Client.Metadata.StateLocation = utils.STORE_LOCAL
	cli.Client.Metadata.K8sDistro = p.Distro

	if _, err1 := control_pkg.InitializeStorageFactory(&cli.Client, true); err1 != nil {
		err = err1
		return
	}

	cli.Client.Metadata.IsHA = true
	cli.Client.Metadata.Region = p.Region
	cli.Client.Metadata.Provider = p.Cloud

	// Return
	ok := true
	errStr := ""

	msg, err := controller.DeleteHACluster(&cli.Client)
	if err != nil {
		ok = false
		errStr = err.Error()
	}

	res = &httpserver.Response{OK: &ok, Errors: &errStr, Response: msg}
	s.logger.Print(msg)

	return
}

// Scaledown implements scaledown.
func (s *httpserversrvc) Scaledown(ctx context.Context, p *httpserver.Metadata) (res *httpserver.Response, err error) {
	cli = &resources.CobraCmd{}
	controller = control_pkg.GenKsctlController()

	cli.Client.Metadata.ClusterName = p.ClusterName
	cli.Client.Metadata.StateLocation = utils.STORE_LOCAL
	cli.Client.Metadata.K8sDistro = p.Distro

	if _, err1 := control_pkg.InitializeStorageFactory(&cli.Client, true); err1 != nil {
		err = err1
		return
	}

	cli.Client.Metadata.IsHA = true
	cli.Client.Metadata.Region = p.Region
	cli.Client.Metadata.Provider = p.Cloud

	cli.Client.Metadata.NoWP = int(*p.NoWp)

	// Return
	ok := true
	errStr := ""

	msg, err := controller.DelWorkerPlaneNode(&cli.Client)
	if err != nil {
		ok = false
		errStr = err.Error()
	}

	res = &httpserver.Response{OK: &ok, Errors: &errStr, Response: msg}
	s.logger.Print(msg)

	return
}

// Scaleup implements scaleup.
func (s *httpserversrvc) Scaleup(ctx context.Context, p *httpserver.Metadata) (res *httpserver.Response, err error) {
	cli = &resources.CobraCmd{}
	controller = control_pkg.GenKsctlController()

	cli.Client.Metadata.ClusterName = p.ClusterName
	cli.Client.Metadata.StateLocation = utils.STORE_LOCAL
	cli.Client.Metadata.K8sDistro = p.Distro

	cli.Client.Metadata.K8sVersion = "1.27.1"
	if _, err1 := control_pkg.InitializeStorageFactory(&cli.Client, true); err1 != nil {
		err = err1
		return
	}

	cli.Client.Metadata.IsHA = true
	cli.Client.Metadata.Region = p.Region
	cli.Client.Metadata.Provider = p.Cloud

	cli.Client.Metadata.WorkerPlaneNodeType = *p.VMSizeWp
	cli.Client.Metadata.NoWP = int(*p.NoWp)

	// Return
	ok := true
	errStr := ""

	msg, err := controller.AddWorkerPlaneNode(&cli.Client)
	if err != nil {
		ok = false
		errStr = err.Error()
	}

	res = &httpserver.Response{OK: &ok, Errors: &errStr, Response: msg}
	s.logger.Print(msg)

	return
}

// GetHealth implements get health.
func (s *httpserversrvc) GetHealth(ctx context.Context) (res *httpserver.Health, err error) {
	abcd := "ksctl server looks good"
	res = &httpserver.Health{Msg: &abcd}
	return
}

// GetClusters implements get clusters.
func (s *httpserversrvc) GetClusters(ctx context.Context) (res *httpserver.Response, err error) {

	cli = &resources.CobraCmd{}

	cli.Client.Metadata.StateLocation = utils.STORE_LOCAL

	if _, err := control_pkg.InitializeStorageFactory(&cli.Client, true); err != nil {
		panic(err)
	}

	var printerTable []cloudController.AllClusterData
	ok := true
	data, err1 := civo_pkg.GetRAWClusterInfos(cli.Client.Storage)
	if err1 != nil {
		err = err1
		ok = false
		return
	}
	printerTable = append(printerTable, data...)

	data, err1 = azure_pkg.GetRAWClusterInfos(cli.Client.Storage)
	if err1 != nil {
		err = err1
		ok = false
		return
	}

	printerTable = append(printerTable, data...)

	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	res = &httpserver.Response{OK: &ok, Errors: &errStr, Response: printerTable}
	s.logger.Print("httpserver.get clusters")
	return
}
