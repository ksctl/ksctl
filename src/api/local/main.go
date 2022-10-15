/*
Kubesimplify
Credit to @kubernetes.io
@maintainer: Dipankar Das <dipankardas0115@gmail.com> , Anurag Kumar <contact.anurag7@gmail.com>
*/

package local

import (
	"fmt"
	"github.com/kubesimplify/Kubesimpctl/src/api/payload"
	"os"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/errors"
	// "sigs.k8s.io/kind/pkg/internal/runtime"
)

// func configOption(rawConfigFlag string, stdin io.Reader) (cluster.CreateOption, error) {
// 	// if not - then we are using a real file
// 	if rawConfigFlag != "-" {
// 		return cluster.CreateWithConfigFile(rawConfigFlag), nil
// 	}
// 	// otherwise read from stdin
// 	raw, err := ioutil.ReadAll(stdin)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "error reading config from stdin")
// 	}
// 	return cluster.CreateWithRawConfig(raw), nil
// }
var (
	kubeconfig = fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/local/", payload.GetUserName())
)

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

	return kubeconfig + clusterName + "/config", nil
}

func CreateCluster(Name string) error {
	provider := cluster.NewProvider(
	// cluster.ProviderWithLogger(logger),  // TODO: try to add these
	// runtime.GetDefault(logger),
	)

	// TODO: multiple node cluster creation
	// withConfig, err := configOption(flags.Config, streams.In)
	// if err != nil {
	// 	return err
	// }
	/**
	 * TODO & DISCUSS
	 * whether to all users to config their clusters or they will specify the node number we will provide cluster
	 * when the HA cluster gate is off then any node number > 1 will result in addtion of line
	 * kind: Cluster
	 * apiVersion: kind.x-k8s.io/v1alpha4
	 * nodes:
	 * - role: control-plane
	 * - role: worker // addition of it
	 *
	 *
	 * if HA
	 * then devise a formula how much controlplane node needed given number of nodes
	 */
	if err := provider.Create(
		Name,
		// withConfig,
		cluster.CreateWithNodeImage("kindest/node:v1.25.2@sha256:9be91e9e9cdf116809841fc77ebdb8845443c4c72fe5218f3ae9eb57fdb4bace"),
		// cluster.CreateWithRetain(flags.Retain),
		// cluster.CreateWithWaitForReady(Wait),
		cluster.CreateWithKubeconfigPath(func() string {
			path, err := createNecessaryConfigs(Name)
			if err != nil {
				panic(err)
			}
			return path
		}()),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	); err != nil {
		return errors.Wrap(err, "failed to create cluster")
	}

	return nil
}

func DeleteCluster(name string) error {
	provider := cluster.NewProvider(
	// cluster.ProviderWithLogger(logger),	// TODO: try to add these
	// runtime.GetDefault(logger),
	)
	_, err := os.ReadFile(kubeconfig + name+"/info")
	if err != nil {
		return fmt.Errorf("NO matching cluster found")
	}

	if err := provider.Delete(name, kubeconfig+name+"/config"); err != nil {
		return fmt.Errorf("FAIL to delete cluster %q", "abcd")
	}
	return nil
}
