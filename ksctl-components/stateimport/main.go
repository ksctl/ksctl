package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
)

var (
	log = logger.NewStructuredLogger(func() int {
		if v, ok := os.LookupEnv("LOG_LEVEL"); !ok {
			return 0
		} else {
			if v == "DEBUG" {
				return -1
			} else {
				return 0
			}
		}
	}(), os.Stdout)

	ctx = context.WithValue(context.Background(), consts.KsctlModuleNameKey, "ksctl-stateimport")
)

type HealthRes struct {
	Status      string
	Version     string
	ServerName  string
	Description string
}

type ImportRes struct {
	Status      string
	Description string
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	_, _e := writeJson(
		w,
		http.StatusOK,
		HealthRes{
			Status:      "OK",
			Version:     "v1alpha1",
			ServerName:  "ksctl-stateimport",
			Description: "It is a Temporary Server Purpose is to import the exorted data out of local storge to inside the kubernetes cluster",
		},
	)
	if _e != nil {
		log.Error("handled the error", "caught", _e)
	}
	log.Success(ctx, "Handled the request")
}

func importHandler(w http.ResponseWriter, r *http.Request) {
	rawData := types.StorageStateExportImport{}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&rawData)
	if err != nil {
		log.Error("failed to decode the req", "Reason", err)
		_, _e := writeJson(
			w,
			http.StatusBadRequest,
			ImportRes{
				Status:      "failed to decode req",
				Description: err.Error(),
			},
		)
		if _e != nil {
			log.Error("handled the error", "caught", _e)
		}
		return
	}
	client := new(types.KsctlClient)

	client.Metadata.StateLocation = consts.StoreK8s
	log.Debug(ctx, "Metadata for Storage", "client.Metadata", client.Metadata)

	if _, ok := helpers.IsContextPresent(ctx, consts.KsctlTestFlagKey); ok {
		_, _e := writeJson(
			w,
			http.StatusOK,
			ImportRes{
				Status:      "OK, Data got transfered",
				Description: "make sure this service is destroyed",
			},
		)
		if _e != nil {
			log.Error("handled the error", "caught", _e)
		}
		return
	}
	_, err = controllers.NewManagerClusterManaged(ctx, log, client)
	if err != nil {
		log.Error("failed to decode the req", "Reason", err)
		_, _e := writeJson(
			w,
			http.StatusBadRequest,
			ImportRes{
				Status:      "failed to init ksctl manager",
				Description: err.Error(),
			},
		)
		if _e != nil {
			log.Error("handled the error", "caught", _e)
		}
		return
	}
	err = client.Storage.Import(&rawData)
	if err != nil {
		log.Error("failed to decode the req", "Reason", err)
		_, _e := writeJson(
			w,
			http.StatusBadRequest,
			ImportRes{
				Status:      "failed in storage import",
				Description: err.Error(),
			},
		)
		if _e != nil {
			log.Error("handled the error", "caught", _e)
		}
		return
	}

	_, _e := writeJson(
		w,
		http.StatusOK,
		ImportRes{
			Status:      "OK, Data got transfered",
			Description: "make sure this service is destroyed",
		},
	)
	if _e != nil {
		log.Error("handled the error", "caught", _e)
	}
	log.Success(ctx, "Handled the request")
}

func writeJson(w http.ResponseWriter, statusCode int, data any) (int, error) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return http.StatusInternalServerError, log.NewError(ctx, "failed to encode the res", "Reason", err)
	}
	return http.StatusOK, nil
}

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("POST /import", importHandler)

	mux.HandleFunc("GET /healthz", healthzHandler)

	log.Success(ctx, "started the ksctl-stateimport", "port", 8080)
	_e := http.ListenAndServe(":8080", mux)
	if _e != nil {
		panic(_e)
	}
}
