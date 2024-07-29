//go:build testing_k8sclient

package k8sclient_test

import (
	"context"
	"github.com/gookit/goutil/dump"
	"github.com/ksctl/ksctl/internal/kubernetes/k8sclient"
	"github.com/ksctl/ksctl/pkg/logger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"testing"
)

var (
	k8s *k8sclient.K8sClient
	ctx = context.TODO()
	log = logger.NewStructuredLogger(-1, os.Stdout)
)

func TestMain(m *testing.M) {
	var err error
	k8s, err = k8sclient.NewK8sClient(nil)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestK8sClient_PodApply(t *testing.T) {
	o := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Image: "helloworld",
				},
			},
		},
	}
	if err := k8s.PodApply(ctx, log, o); err != nil {
		t.Fatalf("PodApply failed: %s", err.Error())
	}
}

func TestK8sClient_PodDelete(t *testing.T) {
	if err := k8s.PodDelete(ctx, log, &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-pod"}}); err != nil {
		t.Fatalf("PodDelete failed: %s", err.Error())
	}
}

func TestK8sClient_NodesList(t *testing.T) {

	if nL, err := k8s.NodesList(ctx, log); err != nil {
		t.Fatalf("NodesList failed: %s", err.Error())
	} else {
		dump.Println(nL)
	}
}
