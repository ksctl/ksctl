package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	control_pkg "github.com/kubesimplify/ksctl/api/controllers"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers"
	"github.com/kubesimplify/ksctl/api/utils"

	azure_pkg "github.com/kubesimplify/ksctl/api/provider/azure"
	civo_pkg "github.com/kubesimplify/ksctl/api/provider/civo"
	cloudController "github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

var (
	cli        *resources.CobraCmd
	controller controllers.Controller
)

type Metadata struct {

	// // desired no of workerplane nodes
	// NoCp *int32
	// // desired no of workerplane nodes
	// NoDs *int32
	// // desired no of workerplane nodes
	// NoMp *int32
	// // virtual machine size for the controlplane
	// VMSizeCp *string
	// // virtual machine size for the datastore
	// VMSizeDs *string
	// // virtual machine size for the loadbalancer
	// VMSizeLb *string

	//https://github.com/go-playground/validator#strings use them to use for cloud to provide contains

	// desired no of workerplane nodes
	NoWp int `json:"nowp"`

	// virtual machine size for the workerplane
	VMSizeWp string `json:"vmsize"`

	// Cluster name
	ClusterName string `json:"clustername"`

	// Region
	Region string `json:"region"`

	// cloud provider
	Cloud string `json:"cloud"`

	// kubernetes distribution
	Distro string `json:"distro"`
}

type Response struct {
	// successful
	OK bool
	// reason of failure
	Errors string
	// response
	Response any
}

func getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "ksctl scaler is HEALTHY!",
	})

}

func scaleUp(context *gin.Context) {
	req := Metadata{}
	// using BindJson method to serialize body with struct
	if err := context.BindJSON(&req); err != nil {
		_ = context.AbortWithError(http.StatusBadRequest, err)
		return
	}

	cli = &resources.CobraCmd{}
	controller = control_pkg.GenKsctlController()

	cli.Client.Metadata.ClusterName = req.ClusterName
	cli.Client.Metadata.StateLocation = utils.STORE_LOCAL
	cli.Client.Metadata.K8sDistro = req.Distro

	cli.Client.Metadata.K8sVersion = "1.27.1"
	if _, err := control_pkg.InitializeStorageFactory(&cli.Client, true); err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	cli.Client.Metadata.IsHA = true
	cli.Client.Metadata.Region = req.Region
	cli.Client.Metadata.Provider = req.Cloud

	cli.Client.Metadata.WorkerPlaneNodeType = req.VMSizeWp
	cli.Client.Metadata.NoWP = int(req.NoWp)

	msg, err := controller.AddWorkerPlaneNode(&cli.Client)
	if err != nil {
		context.JSON(http.StatusInternalServerError, &Response{OK: false, Errors: err.Error()})
	} else {
		context.JSON(http.StatusAccepted, &Response{OK: true, Response: msg})
	}
}

func scaleDown(context *gin.Context) {

	req := Metadata{}
	// using BindJson method to serialize body with struct
	if err := context.BindJSON(&req); err != nil {
		_ = context.AbortWithError(http.StatusBadRequest, err)
		return
	}
	cli = &resources.CobraCmd{}
	controller = control_pkg.GenKsctlController()

	cli.Client.Metadata.ClusterName = req.ClusterName
	cli.Client.Metadata.StateLocation = utils.STORE_LOCAL
	cli.Client.Metadata.K8sDistro = req.Distro

	if _, err := control_pkg.InitializeStorageFactory(&cli.Client, true); err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	cli.Client.Metadata.IsHA = true
	cli.Client.Metadata.Region = req.Region
	cli.Client.Metadata.Provider = req.Cloud

	cli.Client.Metadata.NoWP = int(req.NoWp)

	// Return

	msg, err := controller.DelWorkerPlaneNode(&cli.Client)
	if err != nil {
		context.JSON(http.StatusInternalServerError, &Response{OK: false, Errors: err.Error()})
	} else {
		context.JSON(http.StatusAccepted, &Response{OK: true, Response: msg})
	}
}

func getClusters(context *gin.Context) {

	cli = &resources.CobraCmd{}

	cli.Client.Metadata.StateLocation = utils.STORE_LOCAL

	if _, err := control_pkg.InitializeStorageFactory(&cli.Client, true); err != nil {
		context.JSON(http.StatusInternalServerError, &Response{OK: false, Errors: err.Error()})
		return
	}

	var printerTable []cloudController.AllClusterData

	data, err := civo_pkg.GetRAWClusterInfos(cli.Client.Storage)
	if err != nil {
		context.JSON(http.StatusInternalServerError, &Response{OK: false, Errors: err.Error()})
		return
	}
	printerTable = append(printerTable, data...)

	data, err = azure_pkg.GetRAWClusterInfos(cli.Client.Storage)
	if err != nil {
		context.JSON(http.StatusInternalServerError, &Response{OK: false, Errors: err.Error()})
		return
	}
	printerTable = append(printerTable, data...)

	context.JSON(http.StatusAccepted, &Response{OK: true, Response: printerTable})
}

func main() {

	if err := os.Setenv(utils.KSCTL_TEST_DIR_ENABLED, "/app/ksctl-data"); err != nil {
		panic(err)
	}

	// f, _ := os.Create("gin.log")
	// gin.DefaultWriter = io.MultiWriter(f)

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.GET("/healthz", getHealth)

	r.PUT("/scaleup", scaleUp)

	r.PUT("/scaledown", scaleDown)

	r.GET("/list", getClusters)

	_ = r.Run(":8080")
}

// // CreateHa implements create ha.
// func (s *httpserversrvc) CreateHa(ctx context.Context, p *httpserver.Metadata) (res *httpserver.Response, err error) {
//
// 	cli = &resources.CobraCmd{}
// 	controller = control_pkg.GenKsctlController()
//
// 	cli.Client.Metadata.ClusterName = p.ClusterName
// 	cli.Client.Metadata.StateLocation = utils.STORE_LOCAL
// 	cli.Client.Metadata.K8sDistro = p.Distro
// 	cli.Client.Metadata.K8sVersion = "1.27.1"
//
// 	if _, err := control_pkg.InitializeStorageFactory(&cli.Client, true); err != nil {
// 		panic(err)
// 	}
//
// 	cli.Client.Metadata.Region = p.Region
// 	cli.Client.Metadata.Provider = p.Cloud
//
// 	cli.Client.Metadata.NoCP = int(*p.NoCp)
// 	cli.Client.Metadata.NoDS = int(*p.NoDs)
// 	cli.Client.Metadata.NoWP = int(*p.NoWp)
// 	cli.Client.Metadata.IsHA = true
//
// 	cli.Client.Metadata.ControlPlaneNodeType = *p.VMSizeCp
// 	cli.Client.Metadata.WorkerPlaneNodeType = *p.VMSizeWp
// 	cli.Client.Metadata.LoadBalancerNodeType = *p.VMSizeLb
// 	cli.Client.Metadata.DataStoreNodeType = *p.VMSizeDs
//
// 	// Return
// 	ok := true
// 	errStr := ""
//
// 	msg, err := controller.CreateHACluster(&cli.Client)
// 	if err != nil {
// 		ok = false
// 		errStr = err.Error()
// 	}
//
// 	res = &httpserver.Response{OK: &ok, Errors: &errStr, Response: msg}
// 	s.logger.Print(msg)
//
// 	return
// }
//
// // DeleteHa implements delete ha.
// func (s *httpserversrvc) DeleteHa(ctx context.Context, p *httpserver.Metadata) (res *httpserver.Response, err error) {
//
// 	cli = &resources.CobraCmd{}
// 	controller = control_pkg.GenKsctlController()
//
// 	cli.Client.Metadata.ClusterName = p.ClusterName
// 	cli.Client.Metadata.StateLocation = utils.STORE_LOCAL
// 	cli.Client.Metadata.K8sDistro = p.Distro
//
// 	if _, err1 := control_pkg.InitializeStorageFactory(&cli.Client, true); err1 != nil {
// 		err = err1
// 		return
// 	}
//
// 	cli.Client.Metadata.IsHA = true
// 	cli.Client.Metadata.Region = p.Region
// 	cli.Client.Metadata.Provider = p.Cloud
//
// 	// Return
// 	ok := true
// 	errStr := ""
//
// 	msg, err := controller.DeleteHACluster(&cli.Client)
// 	if err != nil {
// 		ok = false
// 		errStr = err.Error()
// 	}
//
// 	res = &httpserver.Response{OK: &ok, Errors: &errStr, Response: msg}
// 	s.logger.Print(msg)
//
// 	return
// }
//
