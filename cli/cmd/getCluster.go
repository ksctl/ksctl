package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
*/

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kubesimplify/ksctl/api/provider/logger"
	util "github.com/kubesimplify/ksctl/api/provider/utils"
	"github.com/spf13/cobra"
)

type printer struct {
	ClusterName string `json:"cluster_name"`
	Region      string `json:"region"`
	Provider    string `json:"provider"`
}

const (
	ALL    = int(0)
	CIVOC  = int(1)
	LOCALC = int(2)
	AZUREC = int(3)
)

func printUtil(cargo []byte) {
	//TODO: Added Table type display
	log := logger.Logger{}
	log.Print("\n" + string(cargo))
}

// Printer
func Printer(i int) {
	var toBePrinted []printer

	folders, err := os.ReadDir(util.GetPath(util.CLUSTER_PATH, "civo", "managed"))
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range folders {
		if file.IsDir() {
			info := strings.Split(file.Name(), " ")
			toBePrinted = append(toBePrinted, printer{ClusterName: info[0], Region: info[1], Provider: "CIVO (MANAGED)"})
		}
	}

	folders, err = os.ReadDir(util.GetPath(util.CLUSTER_PATH, "civo", "ha"))
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range folders {
		if file.IsDir() {
			info := strings.Split(file.Name(), " ")
			toBePrinted = append(toBePrinted, printer{ClusterName: info[0], Region: info[1], Provider: "CIVO (HA)"})
		}
	}

	folders, err = os.ReadDir(util.GetPath(util.CLUSTER_PATH, "azure", "ha"))
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range folders {
		if file.IsDir() {
			info := strings.Split(file.Name(), " ")
			toBePrinted = append(toBePrinted, printer{ClusterName: info[0], Region: info[2], Provider: "AZURE (HA)"})
		}
	}

	folders, err = os.ReadDir(util.GetPath(util.CLUSTER_PATH, "azure", "managed"))
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range folders {
		if file.IsDir() {
			info := strings.Split(file.Name(), " ")
			toBePrinted = append(toBePrinted, printer{ClusterName: info[0], Region: info[2], Provider: "AZURE (MANAGED)"})
		}
	}

	folders, err = os.ReadDir(util.GetPath(util.CLUSTER_PATH, "local"))
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range folders {
		if file.IsDir() {
			toBePrinted = append(toBePrinted, printer{ClusterName: file.Name(), Region: "N/A", Provider: "local"})
		}
	}

	arr, err := json.MarshalIndent(toBePrinted, "", "  ")
	if err != nil {
		panic(fmt.Errorf("JSON Convertion failed"))
	}
	if len(toBePrinted) == 0 {
		log := logger.Logger{}
		log.Info("No clusters found", "")
	} else {
		printUtil(arr)

	}
}

// viewClusterCmd represents the viewCluster command
var getClusterCmd = &cobra.Command{
	Use:     "get-clusters",
	Aliases: []string{"get"},
	Short:   "Use to get clusters",
	Long: `It is used to view clusters. For example:

ksctl get-clusters `,
	Run: func(cmd *cobra.Command, args []string) {
		Printer(ALL)
		//Printer(CIVOC)
		//Printer(LOCALC)
	},
}

func init() {
	rootCmd.AddCommand(getClusterCmd)
}
