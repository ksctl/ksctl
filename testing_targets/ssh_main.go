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
	user := flag.String("user", "ubuntu", "user name")
	ip := flag.String("ip", "212.2.247.56", "ip address")
	sshKeyPath := flag.String("ssh-key", "/home/dipankar/Onedrive/github/ksctl/testing_targets/abcd", "ssh private key path")
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

	if err := config.Flag(consts.UtilExecWithOutput).
		IPv4(*ip).
		Script(demoScript()).
		FastMode(true).
		SSHExecute(log); err != nil {
		panic(err)
	}

	fmt.Println(config.GetOutput())
}

func demoScript() resources.ScriptCollection {

	collection := helpers.NewScriptCollection()

	collection.Append(resources.Script{
		Name:       "pinging",
		CanRetry:   true,
		MaxRetries: 9,
		ShellScript: `
ping -c 8 www.google.com
`,
		ScriptExecutor: consts.LinuxBash,
	})
	collection.Append(resources.Script{
		Name:     "os-release",
		CanRetry: false,
		ShellScript: `
cat /etc/os-release
`,
		ScriptExecutor: consts.LinuxBash,
	})
	return collection
}
