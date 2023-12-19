package controllers

import (
	"fmt"
	"runtime"
)

func printKubeConfig(path string) {
	key := ""
	box := ""
	switch runtime.GOOS {
	case "windows":
		key = "$Env:KUBECONFIG"
	case "linux", "darwin":
		key = "export KUBECONFIG"
	}

	box = key + "=" + fmt.Sprintf("\"%s\"", path)
	log.Box("KUBECONFIG env var", box)
}
