package main

import (
	"errors"
	"github.com/kubesimplify/ksctl/internal/logger"
	controlPkg "github.com/kubesimplify/ksctl/pkg/controllers"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"os"
)

func main() {
	var ksctl resources.KsctlClient
	ksctl.Logger = new(logger.Logger)

	//{
	//	civoLog := log.WithGroup("civo")
	//	civoLog = civoLog.With("err", nil)
	//	civoLog.Info("Civo started")
	//}

	if _, err := controlPkg.InitializeLoggerFactory(&ksctl, os.Stdout, -1); err != nil {
		panic(err)
	}

	ksctl.Logger.Print("Printed", "type", "demo")
	ksctl.Logger.Debug("Debugged", "type", "demo")
	ksctl.Logger.Error("Errored", "type", "demo")
	ksctl.Logger.Warn("Warned", "type", "demo")
	ksctl.Logger.Success("Successed", "type", "demo")
	ksctl.Logger.Note("Noted", "type", "demo")
	{
		ksctl.Logger.SetModule("civo")
		ksctl.Logger.Success("Created the cluster", "ksctlClient", ksctl)
	}
	{
		ksctl.Logger.SetModule("azure")
		ksctl.Logger.Print("Created the cluster", "ksctlClient", ksctl)
	}

	ksctl.Logger.SetModule("ksctl")
	ksctl.Logger.Error("Failed", "Err", errors.New("fake error"))
}
