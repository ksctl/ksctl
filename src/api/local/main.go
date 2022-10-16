/*
Kubesimplify
Credit to @kubernetes.io
@maintainer: Dipankar Das <dipankardas0115@gmail.com> , Anurag Kumar <contact.anurag7@gmail.com>
*/

package local

import (
	"fmt"
	"os"
	"time"

	"github.com/kubesimplify/Kubesimpctl/src/api/payload"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/errors"
)

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

var (
	kubeconfig = fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/local/", payload.GetUserName())
)

func isPresent(cluster string) bool {
	_, err := os.ReadFile(kubeconfig + cluster + "/info")
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return false
}

func createNecessaryConfigs(clusterName string) (string, error) {
	workingDir := kubeconfig + clusterName
	err := os.Mkdir(workingDir, 0750)
	if err != nil && !os.IsExist(err) {
		return "", err
	}

	_, err = os.Create(workingDir + "/config")
	if err != nil {
		// TODO: if error happens here do clean up the dir created above
		return "", err
	}
	_, err = os.Create(workingDir + "/info")
	if err != nil {
		return "", err
	}

	err = os.WriteFile(
		fmt.Sprintf(workingDir+"/info"),
		[]byte(fmt.Sprintf("%s", clusterName)),
		0640)

	if err != nil {
		return "", err
	}

	return workingDir + "/config", nil
}

func CreateCluster(Name string, nodes int) error {
	//logg := log.Logger{}
	provider := cluster.NewProvider(
	//cluster.ProviderWithLogger(logg), // TODO: try to add these
	//runtime.GetDefault(log),
	)

	withConfig, err := configOption(nodes)
	if err != nil {
		return err
	}
	if isPresent(Name) {
		return fmt.Errorf("DUPLICATE cluster creation")
	}

	Wait := 50 * time.Second
	if err := provider.Create(
		Name,
		withConfig,
		cluster.CreateWithNodeImage("kindest/node:v1.25.2@sha256:9be91e9e9cdf116809841fc77ebdb8845443c4c72fe5218f3ae9eb57fdb4bace"),
		// cluster.CreateWithRetain(flags.Retain),
		cluster.CreateWithWaitForReady(Wait),
		cluster.CreateWithKubeconfigPath(func() string {
			path, err := createNecessaryConfigs(Name)
			if err != nil {
				_ = deleteConfigs(kubeconfig + Name) // for CLEANUP
				panic(err)
			}
			return path
		}()),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	); err != nil {
		_ = deleteConfigs(kubeconfig + Name) // for CLEANUP
		return errors.Wrap(err, "failed to create cluster")
	}

	var abc payload.PrinterKubeconfigPATH
	abc = printer{ClusterName: Name}
	abc.Printer(0)
	return nil
}

type printer struct {
	ClusterName string
}

func (p printer) Printer(a int) {
	switch a {
	case 0:
		fmt.Printf("\nTo use this cluster set this environment variable\n\n")
		fmt.Println(fmt.Sprintf("export KUBECONFIG='/home/%s/.kube/kubesimpctl/config/local/%s/config'", payload.GetUserName(), p.ClusterName))
	case 1:
		fmt.Printf("\nUse the following command to unset KUBECONFIG\n\n")
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
	_, err := os.ReadFile(kubeconfig + name + "/info")
	if err != nil {
		return fmt.Errorf("NO matching cluster found")
	}

	if err := provider.Delete(name, kubeconfig+name+"/config"); err != nil {
		return fmt.Errorf("FAIL to delete cluster %q", "abcd")
	}
	if err := deleteConfigs(kubeconfig + name); err != nil {
		return err
	}
	var abc payload.PrinterKubeconfigPATH
	abc = printer{ClusterName: name}
	abc.Printer(1)
	return nil
}
