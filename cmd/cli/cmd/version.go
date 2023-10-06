package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

const ksctl = `
[38;5;93m _  [38;5;99m    [38;5;63m     [38;5;69m    [38;5;33m _  [38;5;39m  _ 
[38;5;93m| |[38;5;99m __ _[38;5;63m__  [38;5;69m ___[38;5;33m | |_[38;5;39m | |
[38;5;93m| |[38;5;99m/ //[38;5;63m __| [38;5;69m/ __[38;5;33m|| _[38;5;39m_|| |
[38;5;93m|  [38;5;99m < \[38;5;63m__ \[38;5;69m| (__[38;5;33m | |[38;5;39m_ | [38;5;38m|
[38;5;93m|_[38;5;99m|\_\[38;5;63m|___/[38;5;69m \__[38;5;33m_| \[38;5;39m__||_[38;5;38m|
[38;5;93m  [38;5;99m    [38;5;63m    [38;5;69m     [38;5;33m    [38;5;39m    [38;5;38m  
[0m
`

// change this using ldflags
var Version string = "dev"

var BuildDate string = time.Now().String()

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ksctl",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(ksctl)
		fmt.Println("Version:", Version)
		fmt.Println("BuildDate:", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
