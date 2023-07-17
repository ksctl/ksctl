package resources_test

import (
	"testing"

	"github.com/kubesimplify/ksctl/api/k8s_distro/k3s"
	"github.com/kubesimplify/ksctl/api/k8s_distro/kubeadm"
	"github.com/kubesimplify/ksctl/api/provider/azure"
	civo "github.com/kubesimplify/ksctl/api/provider/civo/interfaces"
	"github.com/kubesimplify/ksctl/api/provider/local"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/storage/localstate"
	"github.com/kubesimplify/ksctl/api/storage/remotestate"
)

var (
	testClientSet *resources.ClientSet
)

func TestCloudHandlers(t *testing.T) {
	if testClientSet == nil {
		testClientSet = &resources.ClientSet{}
	}

	testCloudSet := []string{"azure", "civo", "local"}
	for _, cloud := range testCloudSet {
		ret := testClientSet.CloudHandler(cloud)
		switch ret.(type) {
		case *civo.CivoProvider, *azure.AzureProvider, *local.LocalProvider:
			t.Logf("Matched!")
		default:
			t.Fatalf("Mismatched!")
		}
	}
}

func TestDistroHandlers(t *testing.T) {
	if testClientSet == nil {
		testClientSet = &resources.ClientSet{}
	}

	testDistroSet := []string{"k3s", "kubeadm"}
	for _, distro := range testDistroSet {
		ret := testClientSet.DistroHandler(distro)
		switch ret.(type) {
		case *k3s.K3sDistro, *kubeadm.KubeadmDistro:
			t.Logf("Matched!")
		default:
			t.Fatalf("Mismatched!")
		}
	}
}

func TestStateHandlers(t *testing.T) {

	if testClientSet == nil {
		testClientSet = &resources.ClientSet{}
	}

	testSateSet := []string{"local", "remote"}
	for _, state := range testSateSet {
		ret := testClientSet.StateHandler(state)
		switch ret.(type) {
		case *localstate.LocalStorageProvider, *remotestate.RemoteStorageProvider:
			t.Logf("Matched!")
		default:
			t.Fatalf("Mismatched!")
		}
	}
}
