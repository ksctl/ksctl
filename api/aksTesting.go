package main

import (
	"github.com/kubesimplify/ksctl/api/logger"
)

func main() {
	logger := logger.Logger{}
	logger.Msg = "Created the resource"
	logger.Info("demo-ns")
	logger.Info("")

	logger.Msg = "RETRYING ..."
	logger.Warn()

	logger.Msg = "INVALID CREDENTIALS"
	logger.Err()
}
