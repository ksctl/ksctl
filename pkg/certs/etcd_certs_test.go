package certs

import (
	"context"
	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"os"
	"testing"
)

var (
	log logger.Logger = logger.NewStructuredLogger(-1, os.Stdout)
)

func TestGenerateCerts(t *testing.T) {
	if ca, etcd, key, err := GenerateCerts(
		context.WithValue(
			context.TODO(),
			consts.KsctlModuleNameKey,
			"demo"),
		log, []string{"192.168.1.1"}); err != nil {
		t.Fatalf("it shouldn't fail, ca: %v, etcd: %v, key: %v, err: %v\n", ca, etcd, key, err)
	}

	if ca, etcd, key, err := GenerateCerts(
		context.WithValue(context.TODO(), consts.KsctlModuleNameKey, "demo"), log, []string{"192,168.1.1"}); err == nil ||
		len(ca) != 0 ||
		len(etcd) != 0 ||
		len(key) != 0 {
		t.Fatalf("it should fail, ca: %v, etcd: %v, key: %v, err: %v\n", ca, etcd, key, err)
	}
}
