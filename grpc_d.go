package main

import (
	"context"
	"encoding/json"
	"github.com/ksctl/ksctl/ksctl-components/storage"
	"github.com/ksctl/ksctl/pkg/resources"
	"os"
	"time"
)

func main() {
	raw, err := os.ReadFile("/home/dipankar/.ksctl/kubeconfig")
	if err != nil {
		panic(err)
	}

	_err := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		rpcClient, conn, err := storage.NewClient(ctx, string(raw))
		defer cancel()
		defer conn.Close()

		if err != nil {
			return err
		}

		raw, _err := json.Marshal(new(resources.StorageStateExportImport))
		if _err != nil {
			return _err
		}
		return storage.ImportData(ctx, rpcClient, raw)
	}()
	if _err != nil {
		panic(_err)
	}
}
