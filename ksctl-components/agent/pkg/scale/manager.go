package scale

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ksctl/ksctl/ksctl-components/agent/pb"
	"log/slog"

	//control_pkg "github.com/ksctl/ksctl/pkg/controllers"
	//"github.com/ksctl/ksctl/pkg/helpers/consts"
	//"github.com/ksctl/ksctl/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
)

func CallManager(operation string, in *pb.ReqScale) error {
	operation = "DUMMYYYYYY" // remove this to enable the calling cloud autoscaler

	fmt.Println("Working on the cloud provider for auto-scale")
	//client := new(resources.KsctlClient)
	//controller := control_pkg.GenKsctlController()
	//
	//client.Metadata.ClusterName = "example-cluster" // where can it recieve the clustername and other info?
	//client.Metadata.Provider = consts.CloudCivo
	//client.Metadata.Region = ""
	//client.Metadata.NoWP = int(in.ScaleTo)
	//// options for getting data from the configmap as volumemount problem is if updated it is not visible on the deployment volume mount
	//
	//client.Metadata.LogVerbosity = 0
	//if os.Getenv("LOG_LEVEL") == "DEBUG" {
	//	client.Metadata.LogVerbosity = -1
	//}
	//
	//client.Metadata.LogWritter = os.Stdout
	//
	//if err := control_pkg.InitializeStorageFactory(context.WithValue(context.Background(), "USERID", "ksctl-agent"), client); err != nil {
	//	return err
	//}

	switch operation {
	//case "scaleup":
	//	return controller.AddWorkerPlaneNode(client)
	//case "scaledown":
	//	return controller.DelWorkerPlaneNode(client)
	default:
		return WorkOnK8s()
	}
}

func WorkOnK8s() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading in-cluster config: %v\n", err)
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Kubernetes client: %v\n", err)
		return err
	}

	namespace := "ksctl"
	configMapName := "ksctl-state"

	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching ConfigMap: %v\n", err)
		return err
	}

	type CM struct {
		ClusterName string `json:"cluster_name"`
		Region      string `json:"region"`
		Count       int8   `json:"count"`
	}
	slog.Info("[original] ksctl-state.ksctl.cm", "entire", cm.Data)

	fmt.Println("ConfigMap Data:")
	if v, ok := cm.Data["state.json"]; ok {
		slog.Info("ksctl-state.ksctl.cm", "state.json", v)
		var updatedData CM

		if err := json.Unmarshal([]byte(v), &updatedData); err != nil {
			return err
		}
		updatedData.Count++
		raw, err := json.Marshal(updatedData)
		if err != nil {
			return err
		}

		cm.Data["state.json"] = string(raw)

	} else {
		return fmt.Errorf("not found the correct key in the configmap")
	}

	updated, err := clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	slog.Info("[updated] ksctl-state.ksctl.cm", "entire", updated.Data)

	return nil
}
