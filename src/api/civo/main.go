/*
Kubesimplify (c)
@maintainer: Dipankar Das <dipankardas0115@gmail.com> , Anurag Kumar <contact.anurag7@gmail.com>
*/

package civo

import (
	"errors"
	"fmt"
	"github.com/civo/civogo"
)

const (
	ERRORCODE = "0x001"
	RegionLON = "LON1"
	RegionFRA = "FRA1"
	RegionNYC = "NYC1"
)

func FetchAPIKey() string {
	return ""
}

func CreateCluster(regionCode string, clusterName string) string {
	client, err := civogo.NewClient(FetchAPIKey(), regionCode)
	if err != nil {
		panic(err.Error())
	}
	defaultNetwork, err := client.GetDefaultNetwork()
	if err != nil {
		panic(err.Error())
	}
	configK8s := &civogo.KubernetesClusterConfig{Name: clusterName, Region: regionCode, NumTargetNodes: 1, NetworkID: defaultNetwork.ID}

	resp, err := client.NewKubernetesClusters(configK8s)
	if err != nil {
		if errors.Is(err, civogo.DatabaseKubernetesClusterDuplicateError) {
			fmt.Println(fmt.Errorf("DUPLICATE NAME FOUND").Error())
			return ERRORCODE
		}
		if errors.Is(err, civogo.AuthenticationFailedError) {
			fmt.Println(fmt.Errorf("AUTH FAILED").Error())
			return ERRORCODE
		}
		if errors.Is(err, civogo.UnknownError) {
			fmt.Println(fmt.Errorf("UNKNOWN ERR").Error())
			return ERRORCODE
		}
	}
	fmt.Println(resp.Status)
	return resp.ID
}

func DeleteCluster(regionCode string, clusterID string) string {
	client, err := civogo.NewClient(FetchAPIKey(), regionCode)
	if err != nil {
		panic(err.Error())
	}

	cluster, err := client.DeleteKubernetesCluster(clusterID)
	if err != nil {
		return err.Error()
	}
	// remove the KUBECONFIG and related configs
	return string(cluster.Result)
}
