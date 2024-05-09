package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"
	"github.com/ksctl/ksctl/pkg/resources/controllers"

	control_pkg "github.com/ksctl/ksctl/pkg/controllers"
)

var (
	cli        *resources.KsctlClient
	controller controllers.Controller
	dir        = fmt.Sprintf("%s/ksctl-http-logger", os.TempDir())
)

func StartCloud(w http.ResponseWriter) {

	cli = new(resources.KsctlClient)
	controller = control_pkg.GenKsctlController()

	cli.Metadata.ClusterName = "fake"
	cli.Metadata.StateLocation = consts.StoreLocal
	cli.Metadata.K8sDistro = consts.K8sK3s

	cli.Metadata.LogVerbosity = 0
	cli.Metadata.LogWritter = w

	if err := control_pkg.InitializeStorageFactory(context.WithValue(context.Background(), "USERID", "scalar"), cli); err != nil {
		panic(err)
	}
}

func createDummyCivo(w http.ResponseWriter, r *http.Request) {
	StartCloud(w)
	log := logger.NewStructuredLogger(-1, w)
	log.SetPackageName("example-http-logger")

	cli.Metadata.Region = "LON1"
	cli.Metadata.Provider = consts.CloudCivo
	cli.Metadata.ManagedNodeType = "g4s.kube.small"
	cli.Metadata.NoMP = 2

	_ = os.Setenv(string(consts.KsctlCustomDirEnabled), dir)

	if err := controller.CreateManagedCluster(cli); err != nil {
		log.Error("Error Encournted", "Reason", err.Error())
		return
	}

	if err := controller.DeleteManagedCluster(cli); err != nil {
		log.Error("Error Encournted", "Reason", err.Error())
		return
	}
	log.Success("Passed")
}

func main() {
	mapRouteToHandler := map[string]func(w http.ResponseWriter, r *http.Request){
		"/": createDummyCivo,
	}

	if err := os.Setenv(string(consts.KsctlFakeFlag), "1"); err != nil {
		panic(err)
	}
	for k, v := range mapRouteToHandler {
		http.HandleFunc(k, v)
	}

	s := &http.Server{
		Addr: ":8080",
	}

	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}
