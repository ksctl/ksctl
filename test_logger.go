package main

import (
	"github.com/kubesimplify/ksctl/internal/logger"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"os"
)

func main() {
	var ksctl resources.KsctlClient
	ksctl.Logger = &logger.Logger{}

	if err := ksctl.Logger.New(5, os.Stdout); err != nil {
		panic(err)
	}
	{
		ksctl.Logger.AppendPrefix("[block 1]")
		ksctl.Logger.Info(resources.MsgTypeSuccess, "creating")
	}
	{
		ksctl.Logger.ResetSetPrefix("[block 2]")
		ksctl.Logger.Info(resources.MsgTypeError, "creating")
		ksctl.Logger.Debug(resources.MsgTypeError, "creating")
	}
	{
		ksctl.Logger.ResetSetPrefix("[block 3]")
		ksctl.Logger.Info(resources.MsgTypeWarn, "creating")
	}
	ksctl.Logger.Infof(resources.MsgTypeSuccess, "Author: %s", "working correctly")
}
