package kubernetes

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
)

func httpClient(isTokenBased bool, caCert, clientCert, clientKey []byte) (*tls.Config, error) {

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	if isTokenBased {
		tlsConfig := &tls.Config{
			RootCAs: caCertPool,
		}
		return tlsConfig, nil
	}

	cert, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, ksctlErrors.ErrFailedConnectingKubernetesCluster.Wrap(
			log.NewError(kubernetesCtx, "Error loading client certificate and key", "Reason", err),
		)
	}

	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}
	return tlsConfig, nil
}

func ExtractURLAndTLSCerts(kubeconfig, clusterContextName string) (url string, tlsConf *tls.Config, token *string, err error) {

	config, err := clientcmd.Load([]byte(kubeconfig))
	if err != nil {
		return "", nil, nil, ksctlErrors.ErrFailedConnectingKubernetesCluster.Wrap(
			log.NewError(kubernetesCtx, "failed deserializes the contents into Config object", "Reason", err),
		)
	}

	clusterContext := ""
	authContext := ""
	isPresent := false
	log.Print(kubernetesCtx, "searching for current-context", "contextName", clusterContextName)
	if config.CurrentContext != clusterContextName {
		log.Warn(kubernetesCtx, "failed context looking for is not the current one", "expected", clusterContextName, "got", config.CurrentContext)
		log.Print(kubernetesCtx, "using the context which is present in the state for configuration", "stateContext", clusterContextName)
	}

	for ctxK8s, info := range config.Contexts {

		if ctxK8s == clusterContextName {
			isPresent = true
			clusterContext = info.Cluster
			authContext = info.AuthInfo
			log.Print(kubernetesCtx, "Found cluster in kubeconfig",
				"current-context", config.CurrentContext,
				"contexts[...].context.cluster", clusterContext,
				"contexts[...].context.authinfo", authContext,
			)
		}
	}

	if !isPresent {
		return "", nil, nil, ksctlErrors.ErrFailedConnectingKubernetesCluster.Wrap(
			log.NewError(kubernetesCtx, "failed to find the context", "contextName", clusterContextName),
		)
	}

	cluster := config.Clusters[clusterContext]
	kubeapiURL := cluster.Server

	if authContext == "aws" {
		token := config.AuthInfos[authContext].Token
		if len(token) == 0 {
			return "", nil, nil, ksctlErrors.ErrFailedConnectingKubernetesCluster.Wrap(
				log.NewError(kubernetesCtx, "failed to get the token", "Reason", "token is empty"),
			)
		}
		tlsConf, _err := httpClient(true, cluster.CertificateAuthorityData, nil, nil)
		if _err != nil {
			return "", nil, nil, _err
		}

		return kubeapiURL, tlsConf, &token, nil

	} else {
		usr := config.AuthInfos[authContext]

		caCert := cluster.CertificateAuthorityData
		clientCert := usr.ClientCertificateData
		clientKey := usr.ClientKeyData
		if len(caCert) == 0 || len(clientCert) == 0 || len(clientKey) == 0 {
			return "", nil, nil, ksctlErrors.ErrFailedConnectingKubernetesCluster.Wrap(
				log.NewError(kubernetesCtx, "failed to get the tls certs", "Reason", "one of the tls certs is empty"),
			)
		}

		tlsConf, _err := httpClient(false, caCert, clientCert, clientKey)
		if _err != nil {
			return "", nil, nil, _err
		}

		return kubeapiURL, tlsConf, nil, nil
	}
}

func transferData(kubeconfig,
	clusterContextName,
	podName,
	podNs string,
	podPort int,
	v *types.StorageStateExportImport) error {

	url, tlsConf, token, err := ExtractURLAndTLSCerts(kubeconfig, clusterContextName)
	if err != nil {
		return err
	}

	out, err := json.Marshal(v)
	if err != nil {
		return ksctlErrors.ErrFailedConnectingKubernetesCluster.Wrap(
			log.NewError(kubernetesCtx, "failed to marshal the exported stateDocuments", "Reason", err),
		)
	}

	url = fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s:%d/proxy/import", url, podNs, podName, podPort)

	log.Debug(kubernetesCtx, "full url for state transfer", "url", url)

	var (
		resHttp *http.Response
	)

	tr := &http.Transport{
		TLSClientConfig: tlsConf,
	}

	expoBackoff := helpers.NewBackOff(
		10*time.Second,
		1,
		int(consts.CounterMaxWatchRetryCount),
	)
	_err := expoBackoff.Run(
		kubernetesCtx,
		log,
		func() (err error) {

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(out))
			if err != nil {
				return ksctlErrors.ErrFailedConnectingKubernetesCluster.Wrap(
					log.NewError(kubernetesCtx, "failed, client could not create request", "Reason", err),
				)
			}

			if token != nil {
				req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", *token))
			}

			client := &http.Client{Transport: tr, Timeout: 1 * time.Minute}

			resHttp, err = client.Do(req)
			if err != nil {
				return ksctlErrors.ErrFailedConnectingKubernetesCluster.Wrap(
					log.NewError(kubernetesCtx, "failed to connect",
						"Reason", err,
					),
				)
			}
			return nil
		},
		func() bool {
			return resHttp.StatusCode == http.StatusOK
		},
		nil,
		func() error {
			body, _err := io.ReadAll(resHttp.Body)
			if _err != nil {
				return ksctlErrors.ErrFailedConnectingKubernetesCluster.Wrap(
					log.NewError(kubernetesCtx, "status code was 200, but failed to read response",
						"Reason", _err,
					),
				)
			}
			log.Success(kubernetesCtx, "Response of successful state transfer", "StatusCode", resHttp.StatusCode, "Response", string(body))
			return nil
		},
		"Retrying to get valid response from state transfer",
	)
	if _err != nil {
		return _err
	}

	return nil
}
