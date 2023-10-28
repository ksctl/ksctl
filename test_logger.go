package main

import (
	"github.com/kubesimplify/ksctl/internal/logger"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"os"
)

func main() {
	var ksctl resources.KsctlClient
	ksctl.Logger = &logger.Logger{}

	if err := ksctl.Logger.New(3, os.Stdout); err != nil {
		panic(err)
	}
	{
		ksctl.Logger.AppendPrefix("block 1")
		ksctl.Logger.Info(resources.MsgTypeSuccess, "creating")
	}
	{
		ksctl.Logger.ResetPrefix()
		ksctl.Logger.Info(resources.MsgTypeError, "creating")
		ksctl.Logger.Debug(resources.MsgTypeError, "debug reset")
		{
			ksctl.Logger.AppendPrefix("block 2 inner")
			ksctl.Logger.Info(resources.MsgTypeError, "poped")
		}
	}
	{
		ksctl.Logger.ResetPrefix()
		ksctl.Logger.Info(resources.MsgTypeWarn, "creating cdsjcjneciejdsner dfcs", "wcdascdscdsc")
	}
	ksctl.Logger.Infof(resources.MsgTypeSuccess, "Author: %s nice: %v", "working correctly", "nice")
}
