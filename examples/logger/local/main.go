package main

import (
	"os"

	"github.com/kubesimplify/ksctl/pkg/logger"
	"github.com/kubesimplify/ksctl/pkg/resources"
)

func main() {
	var log resources.LoggerFactory = logger.NewDefaultLogger(-1, os.Stdout)
	log.Print("Example", "key", "value")
}
