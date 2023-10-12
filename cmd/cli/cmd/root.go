/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

package cmd

import (
	"fmt"
	"os"
	"time"

	controlPkg "github.com/kubesimplify/ksctl/pkg/controllers"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/resources/controllers"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"

	"github.com/spf13/cobra"
)

var (
	clusterName string
	region      string
	noCP        int
	noWP        int
	noMP        int
	noDS        int
	nodeSizeMP  string
	nodeSizeCP  string
	nodeSizeWP  string
	nodeSizeLB  string
	nodeSizeDS  string
	apps        string
	cni         string
	provider    string
	//storage     string  // Currently only local storage is present
	distro string
	k8sVer string
	cloud  map[int]string
)

var (
	cli        *resources.CobraCmd
	controller controllers.Controller
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ksctl",
	Short: "CLI tool for managing multiple K8s clusters",
	Long: `CLI tool which can manage multiple K8s clusters
from local clusters to cloud provider specific clusters.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	cli = &resources.CobraCmd{}
	controller = controlPkg.GenKsctlController()

	cloud = map[int]string{
		1: string(CLOUD_AWS),
		2: string(CLOUD_AZURE),
		3: string(CLOUD_CIVO),
		4: string(CLOUD_LOCAL),
	}
	cli.Client.Metadata.StateLocation = STORE_LOCAL

	timer := time.Now()
	err := rootCmd.Execute()
	defer cli.Client.Storage.Logger().Print(fmt.Sprintf("⏰  %v\n", time.Since(timer)))
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.Kubesimpctl.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	verboseFlags()

	argsFlags()
}
