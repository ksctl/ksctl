package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	control_pkg "github.com/kubesimplify/ksctl/pkg/controllers"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/resources/controllers"

	azure_pkg "github.com/kubesimplify/ksctl/internal/cloudproviders/azure"
	civo_pkg "github.com/kubesimplify/ksctl/internal/cloudproviders/civo"
	. "github.com/kubesimplify/ksctl/pkg/helpers/consts"
	cloudController "github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
)

var (
	cli        *resources.KsctlClient
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

func scaleUp(ctx *gin.Context) {
	req := Metadata{}
	// using BindJson method to serialize body with struct
	if err := ctx.BindJSON(&req); err != nil {
		_ = ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	cli = new(resources.KsctlClient)
	controller = control_pkg.GenKsctlController()

	cli.Metadata.ClusterName = req.ClusterName
	cli.Metadata.StateLocation = StoreLocal
	cli.Metadata.K8sDistro = KsctlKubernetes(req.Distro)
	cli.Metadata.LogVerbosity = 0
	cli.Metadata.LogWritter = os.Stdout

	cli.Metadata.K8sVersion = "1.27.1"
	if err := control_pkg.InitializeStorageFactory(context.WithValue(context.Background(), "USERID", "scalar"), cli); err != nil {
		_ = ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	cli.Metadata.IsHA = true
	cli.Metadata.Region = req.Region
	cli.Metadata.Provider = KsctlCloud(req.Cloud)

	cli.Metadata.WorkerPlaneNodeType = req.VMSizeWp
	cli.Metadata.NoWP = int(req.NoWp)

	err := controller.AddWorkerPlaneNode(cli)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &Response{OK: false, Errors: err.Error()})
	} else {
		ctx.JSON(http.StatusAccepted, &Response{OK: true, Response: "scaled up"})
	}
}

func scaleDown(ctx *gin.Context) {

	req := Metadata{}
	// using BindJson method to serialize body with struct
	if err := ctx.BindJSON(&req); err != nil {
		_ = ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	cli = new(resources.KsctlClient)
	controller = control_pkg.GenKsctlController()

	cli.Metadata.ClusterName = req.ClusterName
	cli.Metadata.StateLocation = StoreLocal
	cli.Metadata.K8sDistro = KsctlKubernetes(req.Distro)

	cli.Metadata.LogVerbosity = 0
	cli.Metadata.LogWritter = os.Stdout

	if err := control_pkg.InitializeStorageFactory(context.WithValue(context.Background(), "USERID", "scalar"), cli); err != nil {
		_ = ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	cli.Metadata.IsHA = true
	cli.Metadata.Region = req.Region
	cli.Metadata.Provider = KsctlCloud(req.Cloud)

	cli.Metadata.NoWP = int(req.NoWp)

	// Return

	err := controller.DelWorkerPlaneNode(cli)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &Response{OK: false, Errors: err.Error()})
	} else {
		ctx.JSON(http.StatusAccepted, &Response{OK: true, Response: "scaled down"})
	}
}

func getClusters(ctx *gin.Context) {

	cli = new(resources.KsctlClient)

	cli.Metadata.StateLocation = StoreLocal

	cli.Metadata.LogVerbosity = 0
	cli.Metadata.LogWritter = os.Stdout

	if err := control_pkg.InitializeStorageFactory(context.WithValue(context.Background(), "USERID", "scalar"), cli); err != nil {
		ctx.JSON(http.StatusInternalServerError, &Response{OK: false, Errors: err.Error()})
		return
	}

	var printerTable []cloudController.AllClusterData
	data, err := civo_pkg.GetRAWClusterInfos(cli.Storage, cli.Metadata)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &Response{OK: false, Errors: err.Error()})
		return
	}
	printerTable = append(printerTable, data...)

	data, err = azure_pkg.GetRAWClusterInfos(cli.Storage, cli.Metadata)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &Response{OK: false, Errors: err.Error()})
		return
	}
	printerTable = append(printerTable, data...)

	ctx.JSON(http.StatusAccepted, &Response{OK: true, Response: printerTable})
}

func main() {

	if err := os.Setenv(string(KsctlCustomDirEnabled), "app ksctl-data"); err != nil {
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
