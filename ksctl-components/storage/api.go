package storage

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/ksctl/ksctl/api/gen/agent/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/client-go/tools/clientcmd"
)

func httpClient(caCert, clientCert, clientKey []byte) (*tls.Config, error) {

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		fmt.Println("Error loading client certificate and key:", err)
		return nil, err
	}

	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}
	return tlsConfig, nil
}

func ExtractURLAndTLSCerts(kubeconfig string) (url string, tlsConf *tls.Config, err error) {
	config, err := clientcmd.Load([]byte(kubeconfig))

	if err != nil {
		return "", nil, err
	}

	cluster := config.Clusters["kind-kind"]
	usr := config.AuthInfos["kind-kind"]

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

func NewClient(ctx context.Context, kubeconfig string) (pb.KsctlAgentClient, *grpc.ClientConn, error) {

	url, tlsConf, err := ExtractURLAndTLSCerts(kubeconfig)
	if err != nil {
		return nil, nil, err
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConf)),
	}

	// agent.ksctl.svc.cluster.local:8080
	url = fmt.Sprintf("%s/api/v1/namespaces/%s/services/%s:%d/proxy/", url, "ksctl", "agent", 8080)
	fmt.Println(url)

	conn, err := grpc.DialContext(ctx, url, opts...)
	if err != nil {
		return nil, conn, err
	}

	return pb.NewKsctlAgentClient(conn), conn, nil
}

func ImportData(ctx context.Context, client pb.KsctlAgentClient, data []byte) error {
	_, err := client.Storage(ctx, &pb.ReqStore{Operation: pb.StorageOperation_IMPORT, Data: data})
	if err != nil {
		return err
	}
	return nil
}
