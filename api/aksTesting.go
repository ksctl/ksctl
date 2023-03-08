package main

import (
	"github.com/kubesimplify/ksctl/api/logger"
)

func main() {
	// logger := logger.Logger{Verbose: true}
	logger := logger.Logger{}
	logger.Info("Created the resource", "demo-ns")
	logger.Info("Created the resource", "")

	logger.Warn("RETRYING ...")

	logger.Err("INVALID CREDENTIALS")

	logger.Print("Hello")
}
