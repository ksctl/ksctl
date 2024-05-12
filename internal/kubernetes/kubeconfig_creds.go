package kubernetes

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ksctl/ksctl/api/gen/agent/pb"
	"github.com/ksctl/ksctl/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
)

func httpClient(caCert, clientCert, clientKey []byte) (*tls.Config, error) {

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, log.NewError(kubernetesCtx, "Error loading client certificate and key", "Reason", err)
	}

	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}
	return tlsConfig, nil
}

func ExtractURLAndTLSCerts(kubeconfig, clusterContextName string) (url string, tlsConf *tls.Config, err error) {

	config, err := clientcmd.Load([]byte(kubeconfig))
	if err != nil {
		return "", nil, log.NewError(kubernetesCtx, "failed deserializes the contents into Config object", "Reason", err)
	}

	cluster := config.Clusters[clusterContextName]
	usr := config.AuthInfos[clusterContextName]

	kubeapiURL := cluster.Server
	caCert := cluster.CertificateAuthorityData
	clientCert := usr.ClientCertificateData
	clientKey := usr.ClientKeyData

	tlsConf, _err := httpClient(caCert, clientCert, clientKey)
	if _err != nil {
		return "", nil, _err
	}
	return kubeapiURL, tlsConf, nil
}

func transferData(kubeconfig, clusterContextName string, v *types.StorageStateExportImport) error {

	url, tlsConf, err := ExtractURLAndTLSCerts(kubeconfig, clusterContextName)
	if err != nil {
		return err
	}

	out, err := json.Marshal(v)
	if err != nil {
		return log.NewError(kubernetesCtx, "failed to marshal the exported stateDocuments", "Reason", err)
	}

	url = fmt.Sprintf("%s/api/v1/namespaces/%s/services/%s:%d/proxy/", url, KSCTL_SYS_NAMESPACE, "ksctl-storeimporter", 80)

	req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer(out))
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		os.Exit(1)
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConf,
	}

	client := &http.Client{Transport: tr, Timeout: 1 * time.Minute}

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Response:", string(body))

	return nil
}

func ImportData(ctx context.Context, client pb.KsctlAgentClient, data []byte) error {
	_, err := client.Storage(ctx, &pb.ReqStore{Operation: pb.StorageOperation_IMPORT, Data: data})
	if err != nil {
		return err
	}
	return nil
}
