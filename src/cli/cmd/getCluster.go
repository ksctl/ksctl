package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
*/

import (
	"encoding/json"
	"fmt"
	"github.com/kubesimplify/Kubesimpctl/src/api/civo"
	"github.com/kubesimplify/Kubesimpctl/src/api/local"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"strings"
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
)

func printUtil(cargo []byte) {
	//TODO: Added Table type display
	fmt.Println(string(cargo))
}

func Printer(i int) {
	var toBePrinted []printer

	files, err := ioutil.ReadDir(civo.KUBECONFIG_PATH)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			info := strings.Split(file.Name(), "-")
			toBePrinted = append(toBePrinted, printer{ClusterName: info[0], Region: info[1], Provider: "civo"})
		}
	}

	files, err = ioutil.ReadDir(local.KUBECONFIG_PATH)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			toBePrinted = append(toBePrinted, printer{ClusterName: file.Name(), Region: "N/A", Provider: "local"})
		}
	}

	arr, err := json.MarshalIndent(toBePrinted, "", "  ")
	if err != nil {
		panic(fmt.Errorf("JSON Convertion failed"))
	}
	printUtil(arr)
}

// viewClusterCmd represents the viewCluster command
var getClusterCmd = &cobra.Command{
	Use:     "get-clusters",
	Aliases: []string{"get"},
	Short:   "Use to get clusters",
	Long: `It is used to view clusters. For example:

kubesimpctl get-clusters `,
	Run: func(cmd *cobra.Command, args []string) {
		Printer(ALL)
		//Printer(CIVOC)
		//Printer(LOCALC)
	},
}

func init() {
	rootCmd.AddCommand(getClusterCmd)
}
