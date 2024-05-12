package main

import (
	"context"
	"encoding/json"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
	"net/http"
	"os"
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

	ctx = context.WithValue(context.Background(), consts.ContextModuleNameKey, "ksctl-storageimporter")
)

func writeJson(w http.ResponseWriter, statusCode int, data any) (int, error) {
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return http.StatusInternalServerError, log.NewError(ctx, "failed to encode the res", "Reason", err)
	}
	return http.StatusOK, nil
}

func main() {

	mux := http.NewServeMux()
	port := os.Getenv("PORT")

	mux.HandleFunc("POST /import", func(w http.ResponseWriter, r *http.Request) {
		rawData := types.StorageStateExportImport{}

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&rawData)
		if err != nil {
			log.Error(ctx, "failed to decode the req", "Reason", err)
			_, _e := writeJson(
				w,
				http.StatusBadRequest,
				struct {
					Msg string
				}{},
			)
			if _e != nil {
				log.Error(ctx, "handled the error", "caught", _e)
			}
			return
		}
		log.Debug(ctx, "testing", "rawData", rawData)
		log.Success(ctx, "Handled the request")
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _e := writeJson(
			w,
			http.StatusOK,
			struct {
				Status      string
				Version     string
				ServerName  string
				Description string
			}{
				Status:      "OK",
				Version:     "v1alpha1",
				ServerName:  "ksctl-storageimporter",
				Description: "It is a Temporary Server Purpose is to import the exorted data out of local storge to inside the kubernetes cluster",
			},
		)
		if _e != nil {
			log.Error(ctx, "handled the error", "caught", _e)
		}
		log.Success(ctx, "Handled the request")
	})

	log.Success(ctx, "started the ksctl-storageimport", "port", port)
	_e := http.ListenAndServe(":"+port, mux)
	if _e != nil {
		panic(_e)
	}
}
