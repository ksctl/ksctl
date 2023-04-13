package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const ksctl = `
 _             _   _ 
| | _____  ___| |_| |
| |/ / __|/ __| __| |
|   <\__ \ (__| |_| |
|_|\_\___/\___|\__|_|
`

// change this using ldflags
var Version string = "dev"

var BuildDate string

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ksctl",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ksctl)
		fmt.Println("Version:", Version)
		fmt.Println("BuildDate:", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
