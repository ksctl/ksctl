package main

import (
	"errors"
	"github.com/kubesimplify/ksctl/internal/logger"
	controlPkg "github.com/kubesimplify/ksctl/pkg/controllers"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"log/slog"
	"os"
)

func main() {
	var ksctl resources.KsctlClient
	ksctl.Logger = new(logger.Logger)

	///////////////////////////////
	// Question: should we use the group attribute????
	{
		log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: false,
			Level:     slog.LevelDebug,
		}))
		civoLog := log.WithGroup("civo")
		civoLog = civoLog.With("err", nil)
		civoLog.Info("Civo started")
	}
	///////////////////////////////

	if _, err := controlPkg.InitializeLoggerFactory(&ksctl, os.Stdout, 0); err != nil {
		panic(err)
	}

	ksctl.Logger.SetModule("ksctl")
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
		ksctl.Logger.Warn("ENV credentials not set")
		ksctl.Logger.Print("Created the cluster", "ksctlClient", ksctl)
	}

	ksctl.Logger.SetModule("ksctl")
	ksctl.Logger.Error("Failed", "Err", errors.New("fake error"))
}
