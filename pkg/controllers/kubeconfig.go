package controllers

import (
	"fmt"
	"runtime"

	"github.com/ksctl/ksctl/pkg/types"
)

func printKubeConfig(log types.LoggerFactory, path string) {
	key := ""
	box := ""
	switch runtime.GOOS {
	case "windows":
		key = "$Env:KUBECONFIG"
	case "linux", "darwin":
		key = "export KUBECONFIG"
	}

	box = key + "=" + fmt.Sprintf("\"%s\"", path)
	log.Box(controllerCtx, "KUBECONFIG env var", box)
}
