package main

import (
	"flag"
	"os"

	"github.com/kubesimplify/ksctl/pkg/helpers"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/logger"
	"github.com/kubesimplify/ksctl/pkg/resources"

	localstate "github.com/kubesimplify/ksctl/internal/storage/local"
)

// NOTE: example command to run it
// go run ssh_main.go -user=root -ip=74.220.18.58 -ssh-key=/home/dipankar/civo-test

func main() {
	user := flag.String("user", "root", "user name")
	ip := flag.String("ip", "", "ip address")
	sshKeyPath := flag.String("ssh-key", "", "ssh private key path")

	flag.Parse()

	var config helpers.SSHCollection = new(helpers.SSHPayload)
	config.Username(*user)
	config.LocPrivateKey(*sshKeyPath)

	var storage resources.StorageFactory = localstate.InitStorage()
	var log resources.LoggerFactory = logger.NewDefaultLogger(-1, os.Stdout)
	log.SetPackageName("ksctl-ssh")

	if err := config.Flag(consts.UtilExecWithOutput).IPv4(*ip).Script(demoScript()).FastMode(true).SSHExecute(storage, log); err != nil {
		panic(err)
	}

	log.Print(config.GetOutput())
}

func demoScript() string {
	return `#!/bin/bash
ping -c 8 www.google.com
`
}
