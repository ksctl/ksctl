/*
Kubesimplify
Credit to @kubernetes.io, @kind
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

package local

import (
	"fmt"
	"log"
	"os"
	"time"

	util "github.com/kubesimplify/ksctl/api/utils"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/errors"
)

// func getKubeconfig(params ...string) string {
// 	var ret strings.Builder

// 	if runtime.GOOS == "windows" {
// 		ret.WriteString(fmt.Sprintf("%s\\.ksctl\\config\\local", util.GetUserName()))
// 		for _, item := range params {
// 			ret.WriteString("\\" + item)
// 		}
// 	} else {
// 		ret.WriteString(fmt.Sprintf("%s/.ksctl/config/local", util.GetUserName()))
// 		for _, item := range params {
// 			ret.WriteString("/" + item)
// 		}
// 	}
// 	return ret.String()
// }

// func GetPath(params ...string) string {
// 	return getKubeconfig(params...)
// }

func generateConfig(noWorker, noControl int) ([]byte, error) {
	if noWorker >= 0 && noControl == 0 {
		return nil, fmt.Errorf("invalid config request control node cannot be 0")
	}
	var config string
	config += `---
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
`
	for noControl > 0 {
		config += `- role: control-plane
`
		noControl--
	}

	for noWorker > 0 {
		config += `- role: worker
`
		noWorker--
	}
	config += `...`

	return []byte(config), nil
}

// TODO: Added option to add Nginx Ingress
func configOption(noOfNodes int) (cluster.CreateOption, error) {

	if noOfNodes < 1 {
		return nil, fmt.Errorf("invalid config request control node cannot be 0")
	}
	if noOfNodes == 1 {
		var config string
		config += `---
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
...`
		return cluster.CreateWithRawConfig([]byte(config)), nil
	}
	//control := noOfNodes / 2 // derive the math
	control := 1
	worker := noOfNodes - control
	raw, err := generateConfig(worker, control)
	if err != nil {
		return nil, fmt.Errorf("ERR in node config generation")
	}

	return cluster.CreateWithRawConfig(raw), nil
}

func isPresent(cluster string) bool {
	_, err := os.ReadFile(util.GetPath(util.OTHER_PATH, "local", cluster, "info"))
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func createNecessaryConfigs(clusterName string) (string, error) {
	err := os.Mkdir(util.GetPath(util.OTHER_PATH, "local", clusterName), 0750)
	if err != nil && !os.IsExist(err) {
		return "", err
	}

	_, err = os.Create(util.GetPath(util.OTHER_PATH, "local", clusterName, "config"))
	if err != nil {
		// TODO: if error happens here do clean up the dir created above
		return "", err
	}
	_, err = os.Create(util.GetPath(util.OTHER_PATH, "local", clusterName, "info"))
	if err != nil {
		return "", err
	}

	err = os.WriteFile(
		fmt.Sprintf(util.GetPath(util.OTHER_PATH, "local", clusterName, "info")),
		[]byte(fmt.Sprintf("%s", clusterName)),
		0640)

	if err != nil {
		return "", err
	}

	return util.GetPath(util.OTHER_PATH, "local", clusterName, "config"), nil
}

func ClusterInfoInjecter(clusterName string, noOfNodes int) util.LocalProvider {
	spec := util.LocalProvider{
		ClusterName: clusterName,
		HACluster:   false,
		Spec: util.Machine{
			ManagedNodes: noOfNodes,
			Disk:         "",
			Mem:          "0M",
			Cpu:          "1m",
		},
	}
	return spec
}

func CreateCluster(localConfig util.LocalProvider) error {

	provider := cluster.NewProvider(
	//cluster.ProviderWithLogger(logg), // TODO: try to add these
	//runtime.GetDefault(log),
	)

	withConfig, err := configOption(localConfig.Spec.ManagedNodes)
	if err != nil {
		return err
	}
	if isPresent(localConfig.ClusterName) {
		return fmt.Errorf("🚩 DUPLICATE cluster creation")
	}

	Wait := 50 * time.Second
	if err := provider.Create(
		localConfig.ClusterName,
		withConfig,
		cluster.CreateWithNodeImage("kindest/node:v1.25.2@sha256:9be91e9e9cdf116809841fc77ebdb8845443c4c72fe5218f3ae9eb57fdb4bace"),
		// cluster.CreateWithRetain(flags.Retain),
		cluster.CreateWithWaitForReady(Wait),
		cluster.CreateWithKubeconfigPath(func() string {
			path, err := createNecessaryConfigs(localConfig.ClusterName)
			if err != nil {
				log.Println("[ERR] Cannot continue 😢")
				_ = DeleteCluster(localConfig.ClusterName)
				panic(err)
			}
			return path
		}()),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	); err != nil {
		log.Println("[ERR] Cannot continue 😢")
		_ = DeleteCluster(localConfig.ClusterName)
		return errors.Wrap(err, "failed to create cluster")
	}

	var printKubeconfig util.PrinterKubeconfigPATH
	printKubeconfig = printer{ClusterName: localConfig.ClusterName}
	printKubeconfig.Printer(false, 0)

	log.Println("Created your local cluster!!🥳 🎉 ")

	return nil
}

func (p printer) Printer(ha bool, a int) {
	switch a {
	case 0:
		fmt.Printf("\n\033[33;40mTo use this cluster set this environment variable\033[0m\n\n")
		fmt.Println(fmt.Sprintf("export KUBECONFIG='%s'", util.GetPath(util.OTHER_PATH, "local", p.ClusterName, "config")))
	case 1:
		fmt.Printf("\n\033[33;40mUse the following command to unset KUBECONFIG\033[0m\n\n")
		fmt.Println(fmt.Sprintf("unset KUBECONFIG"))
	}
	fmt.Println()
}

func deleteConfigs(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	return nil
}

func DeleteCluster(name string) error {
	provider := cluster.NewProvider(
	// cluster.ProviderWithLogger(logger),	// TODO: try to add these
	// runtime.GetDefault(logger),
	)
	_, err := os.ReadFile(util.GetPath(util.CLUSTER_PATH, "local", name, "info"))
	if err != nil {
		return fmt.Errorf("NO matching cluster found")
	}

	if err := provider.Delete(name, util.GetPath(util.CLUSTER_PATH, "local", name, "config")); err != nil {
		return fmt.Errorf("FAIL to delete cluster %q", err)
	}
	if err := deleteConfigs(util.GetPath(util.CLUSTER_PATH, "local", name)); err != nil {
		return err
	}
	var printKubeconfig util.PrinterKubeconfigPATH
	printKubeconfig = printer{ClusterName: name}
	printKubeconfig.Printer(false, 1)
	return nil
}

type printer struct {
	ClusterName string
}

// SwitchContext TODO: Add description
func SwitchContext(clusterName string) error {
	if isPresent(clusterName) {
		// TODO: ISSUE #5
		var printKubeconfig util.PrinterKubeconfigPATH
		printKubeconfig = printer{ClusterName: clusterName}
		printKubeconfig.Printer(false, 0)
		return nil
	}
	return fmt.Errorf("ERR Cluster not found")
}
