package common

import (
	"fmt"
	"github.com/ksctl/ksctl/pkg/logger"
	"runtime"
)

func printKubeConfig(log logger.Logger, path string) {
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
