/*
Kubesimplify
Credit to @kubernetes.io
@maintainer: Dipankar Das <dipankardas0115@gmail.com> , Anurag Kumar <contact.anurag7@gmail.com>
*/

package local

import (
	"fmt"

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

func CreateCluster(Name, ImageName string) error {
	provider := cluster.NewProvider(
		// cluster.ProviderWithLogger(logger),
		// runtime.GetDefault(logger),
		)

	// withConfig, err := configOption(flags.Config, streams.In)
	// if err != nil {
	// 	return err
	// }
	if err := provider.Create(
		Name,
		// withConfig,
		cluster.CreateWithNodeImage(ImageName),
		// cluster.CreateWithRetain(flags.Retain),
		// cluster.CreateWithWaitForReady(Wait),
		cluster.CreateWithKubeconfigPath("./config.json"),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
		); err != nil {
		return errors.Wrap(err, "failed to create cluster")
		}

		return nil
}

func DeleteCluster(name, kubeconfig string) error {
	provider := cluster.NewProvider(
		// cluster.ProviderWithLogger(logger),
		// runtime.GetDefault(logger),
		)
	if err := provider.Delete(name, kubeconfig); err != nil {
		return fmt.Errorf("failed to delete cluster %q", "abcd")
	}
	return nil
}


func DockerHandler() string {
	return fmt.Sprintln("Local K8s Called!")
}
