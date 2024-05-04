package storage

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/ksctl/ksctl/api/gen/agent/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	cluster := config.Clusters["kind-test-e2e-local"] // TODO: change and get it from clusterName
	usr := config.AuthInfos["kind-test-e2e-local"]    // TODO: change and get it from clusterName

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

	_, _, err := ExtractURLAndTLSCerts(kubeconfig)
	//url, tlsConf, err := ExtractURLAndTLSCerts(kubeconfig)
	if err != nil {
		return nil, nil, err
	}

	urlSvr := fmt.Sprintf("127.0.0.1:8001/api/v1/namespaces/ksctl/services/agent/proxy/grpc")
	opts := []grpc.DialOption{
		//grpc.WithTransportCredentials(credentials.NewTLS(tlsConf)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// agent.ksctl.svc.cluster.local:80

	conn, err := grpc.DialContext(ctx, urlSvr, opts...)
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
