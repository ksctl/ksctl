package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"
)

// NOTE: example command to run it
// go run ssh_main.go -user=root -ip=74.220.18.58 -ssh-key=/home/dipankar/civo-test

func main() {
	user := flag.String("user", "", "user name")
	ip := flag.String("ip", "", "ip address")
	sshKeyPath := flag.String("ssh-key", "", "ssh private key path")
	flag.Parse()

	fmt.Printf("%s, %s, %s\n", *user, *ip, *sshKeyPath)
	r, err := os.ReadFile(*sshKeyPath)
	if err != nil {
		panic(err)
	}

	var config helpers.SSHCollection = new(helpers.SSHPayload)
	config.Username(*user)
	config.PrivateKey(string(r))

	var log resources.LoggerFactory = logger.NewDefaultLogger(0, os.Stdout)
	log.SetPackageName("ksctl-ssh")

	if err := config.Flag(consts.UtilExecWithOutput).IPv4(*ip).Script(demoScript()).FastMode(true).SSHExecute(log); err != nil {
		panic(err)
	}

	log.Print(config.GetOutput())
}

func demoScript() string {
	return `#!/bin/bash
ping -c 8 www.google.com
cat /etc/os-release
`
}
