/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

package cmd

import (
	"os"

	control_pkg "github.com/kubesimplify/ksctl/api/controllers"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers"
	"github.com/kubesimplify/ksctl/api/utils"

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
	controller = control_pkg.GenKsctlController()

	cloud = map[int]string{
		1: utils.CLOUD_AWS,
		2: utils.CLOUD_AZURE,
		3: utils.CLOUD_CIVO,
		4: utils.CLOUD_LOCAL,
	}
	cli.Client.Metadata.StateLocation = utils.STORE_LOCAL

	err := rootCmd.Execute()
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
