package civo

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/kubesimplify/ksctl/api/logger"
	util "github.com/kubesimplify/ksctl/api/utils"
	"github.com/stretchr/testify/assert"
)

func TestFetchAPIKey(T *testing.T) {
	logging := logger.Logger{}

	apikey := fetchAPIKey(logging)

	if fmt.Sprintf("%T", apikey) != "string" || len(apikey) != 0 {
		T.Fatalf("Invalid Return type or APIKey already present")
	}
}

func TestIsValidNodeSize(T *testing.T) {
	validSizes := []string{"g4s.kube.xsmall", "g4s.kube.small", "g4s.kube.medium", "g4s.kube.large", "g4p.kube.small", "g4p.kube.medium", "g4p.kube.large", "g4p.kube.xlarge", "g4c.kube.small", "g4c.kube.medium", "g4c.kube.large", "g4c.kube.xlarge", "g4m.kube.small", "g4m.kube.medium", "g4m.kube.large", "g4m.kube.xlarge"}
	testData := validSizes[rand.Int()%len(validSizes)]
	assert.Equalf(T, true, isValidSizeManaged(testData), "Returns False for valid size")

	assert.Equalf(T, false, isValidSizeManaged("abcd"), "Returns True for invalid node size")
	assert.Equalf(T, false, isValidSizeManaged("kube.small"), "Returns True for invalid node size")
	assert.Equalf(T, false, isValidSizeManaged("g4s.k3s.small"), "Returns True for invalid node size")
}

//TODO: Test ClusterInfoInjecter()

//TODO: Test kubeconfigDeleter()

//Testing of deleteClusterWithID() and DeleteCluster() and CreateCluster() [TODO Need to be done]

func setup() {
	err := os.MkdirAll(util.GetPath(util.CLUSTER_PATH, "civo", "managed"), 0750)
	if err != nil {
		return
	}
}

func TestIsPresent(t *testing.T) {
	setup()
	t.Cleanup(func() {
		_ = os.RemoveAll(util.GetPath(util.CLUSTER_PATH, "civo"))
	})

	present := isPresent("managed", "demo", "LON1")
	assert.Equal(t, false, present, "with no clusters returns true! (false +ve)")
	err := os.Mkdir(util.GetPath(util.CLUSTER_PATH, "civo", "managed", "demo LON1"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Create(util.GetPath(util.CLUSTER_PATH, "civo", "managed", "demo LON1", "info.json"))
	if err != nil {
		t.Fatal(err)
	}
	present = isPresent("managed", "demo", "LON1")
	assert.Equal(t, true, present, "Failed to detect the cluster (false -ve)")
}

func TestClusterInfoInjecterManagedType(t *testing.T) {
	clusterName := "xYz"
	region := "aBc"
	nodeSize := "k3s"
	logging := logger.Logger{}
	abcd := ClusterInfoInjecter(logging, clusterName, region, nodeSize, 1, "", "")
	worker := CivoProvider{
		ClusterName: clusterName,
		Region:      region,
		Spec:        util.Machine{Disk: nodeSize, ManagedNodes: 1},
		APIKey:      fetchAPIKey(logging),
		Application: "Traefik-v2-nodeport,metrics-server", // EXPLICITLY mentioned the expected data
		CNIPlugin:   "flannel",
	}
	if worker != abcd {
		t.Fatalf("Base check failed")
	}
}

// testing civo components

var (
	civoOperator HACollection
)

func InitTesting_HA(t *testing.T) {
	civoOperator = &HAType{
		Client:        nil,
		NodeSize:      "",
		ClusterName:   "demo",
		DiskImgID:     "id",
		DBFirewallID:  "",
		LBFirewallID:  "",
		CPFirewallID:  "",
		WPFirewallID:  "",
		NetworkID:     "",
		SSHID:         "",
		Configuration: nil,
		SSH_Payload:   nil,
	}
}

func TestSwitchContext(t *testing.T) {

	t.Cleanup(func() {
		_ = os.RemoveAll(util.GetPath(util.OTHER_PATH, "civo"))
	})

	if err := os.MkdirAll(util.GetPath(util.CLUSTER_PATH, "civo", "managed", "demo-1 FRA1"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(util.GetPath(util.CLUSTER_PATH, "civo", "ha", "demo-2 LON1"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(util.GetPath(util.CLUSTER_PATH, "civo", "ha", "demo-2 LON1", "info.json"), []byte("{}"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(util.GetPath(util.CLUSTER_PATH, "civo", "managed", "demo-1 FRA1", "info.json"), []byte("{}"), 0755); err != nil {
		t.Fatal(err)
	}

	civoOperator := CivoProvider{
		ClusterName: "demo",
		Region:      "Abcd",
	}
	if err := civoOperator.SwitchContext(); err == nil {
		t.Fatalf("Passed when their is no matching cluster")
	}
	civoOperator.ClusterName = "demo-1"
	civoOperator.Region = "FRA1"
	civoOperator.HACluster = false

	if err := civoOperator.SwitchContext(); err != nil {
		t.Fatalf("Failed in switching context to %v\nError: %v\n", civoOperator, err)
	}

	civoOperator.ClusterName = "demo-2"
	civoOperator.Region = "LON1"
	civoOperator.HACluster = true

	if err := civoOperator.SwitchContext(); err != nil {
		t.Fatalf("Failed in switching context to %v\nError: %v\n", civoOperator, err)
	}
}

func TestUploadSSHKey(t *testing.T) {}

func TestCreateDatabase(t *testing.T) {}

func TestCreateLoadbalancer(t *testing.T) {}

func TestCreateControlPlane(t *testing.T) {}

func TestCreateWorkerNode(t *testing.T) {}
