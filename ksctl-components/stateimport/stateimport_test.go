package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	"github.com/ksctl/ksctl/pkg/types/storage"
	"gotest.tools/v3/assert"
)

func TestMain(m *testing.M) {
	ctx = context.WithValue(
		ctx,
		consts.KsctlTestFlagKey,
		"true",
	)

	m.Run()
}

func TestImport(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(importHandler))

	t.Run("GET request to /import with invalid request body", func(t *testing.T) {

		_b := []byte{0, 0, 0, 0}

		resp, err := http.Post(svr.URL, "Content-Type: application/json", bytes.NewBuffer(_b))
		t.Cleanup(func() {
			resp.Body.Close()
		})

		if err != nil {
			t.Fatalf("failed in /import, err: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 but got: %s", resp.Status)
		}

		if v := resp.Header.Get("Content-Type"); v != "application/json" {
			t.Fatalf("Content-Type == application/json but got: %s", v)
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed in /import, err: %v", err)
		}
		v := ImportRes{}

		if _err := json.Unmarshal(b, &v); _err != nil {
			t.Fatalf("failed in /import, err: %v", _err)
		}

		assert.DeepEqual(t, v.Status, "failed to decode req")

	})

	t.Run("GET request to /import", func(t *testing.T) {
		payload := &types.StorageStateExportImport{
			Clusters: []*storage.StorageDocument{
				{},
			},
		}

		_b, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed in prepating the request body err: %v", err)
		}

		resp, err := http.Post(svr.URL, "Content-Type: application/json", bytes.NewBuffer(_b))
		t.Cleanup(func() {
			resp.Body.Close()
		})
		if err != nil {
			t.Fatalf("failed in /import, err: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 but got: %s", resp.Status)
		}

		if v := resp.Header.Get("Content-Type"); v != "application/json" {
			t.Fatalf("Content-Type == application/json but got: %s", v)
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed in /import, err: %v", err)
		}
		v := ImportRes{}

		if _err := json.Unmarshal(b, &v); _err != nil {
			t.Fatalf("failed in /import, err: %v", _err)
		}

		assert.DeepEqual(t, v,
			ImportRes{
				Status:      "OK, Data got transfered",
				Description: "make sure this service is destroyed",
			},
		)

	})

}

func TestHealthz(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(healthzHandler))

	t.Run("GET resquest to /healthz", func(t *testing.T) {
		resp, err := http.Get(svr.URL)
		t.Cleanup(func() {
			resp.Body.Close()
		})
		if err != nil {
			t.Fatalf("failed in /healthz, err: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 but got: %s", resp.Status)
		}

		if v := resp.Header.Get("Content-Type"); v != "application/json" {
			t.Fatalf("Content-Type == application/json but got: %s", v)
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed in /healthz, err: %v", err)
		}
		v := HealthRes{}

		if _err := json.Unmarshal(b, &v); _err != nil {
			t.Fatalf("failed in /healthz, err: %v", _err)
		}

		assert.DeepEqual(t, v, HealthRes{
			Status:      "OK",
			Version:     "v1alpha1",
			ServerName:  "ksctl-stateimport",
			Description: "It is a Temporary Server Purpose is to import the exorted data out of local storge to inside the kubernetes cluster",
		})

	})

}
