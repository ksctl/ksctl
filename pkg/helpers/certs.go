package helpers

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"github.com/ksctl/ksctl/pkg/resources"
)

// NOTE: this go is refering to https://shaneutt.com/blog/golang-ca-and-signed-cert-go/

func extractBuffer(buffer *bytes.Buffer) string {
	return buffer.String()
}

func GenerateCerts(log resources.LoggerFactory, etcdMemPrivAddr []string) (caCert string, etcdCert string, etcdKey string, err error) {

	var validIPAddresses []net.IP = []net.IP{net.IPv4(127, 0, 0, 1)}
	for _, ip := range etcdMemPrivAddr {
		if val := net.ParseIP(string(ip)); val != nil {
			validIPAddresses = append(validIPAddresses, val)
		} else {
			return "", "", "", log.NewError("invalid ip address")
		}
	}
	log.Debug("Etcd Members private ip", "ips", validIPAddresses)

	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			CommonName: "etcd cluster",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(2, 0, 0), // for 2 years
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", "", log.NewError(err.Error())
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return "", "", "", log.NewError(err.Error())
	}

	caPEM := new(bytes.Buffer)
	if err := pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}); err != nil {
		return "", "", "", log.NewError(err.Error())
	}

	caCert = extractBuffer(caPEM)
	log.Debug("CA CERTIFICATE", "ca.crt", caCert)

	caPrivKeyPEM := new(bytes.Buffer)
	if err := pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	}); err != nil {
		return "", "", "", log.NewError(err.Error())
	}

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName: "etcd",
		},
		IPAddresses:  validIPAddresses,
		DNSNames:     []string{"localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", "", log.NewError(err.Error())
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return "", "", "", log.NewError(err.Error())
	}

	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		return "", "", "", log.NewError(err.Error())
	}

	certPrivKeyPEM := new(bytes.Buffer)
	if err := pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	}); err != nil {
		return "", "", "", log.NewError(err.Error())
	}

	etcdCert = extractBuffer(certPEM)
	etcdKey = extractBuffer(certPrivKeyPEM)

	log.Debug("ETCD CERTIFICATE", "etcd.crt", etcdCert)
	log.Debug("ETCD KEY", "key.pem", etcdKey)
	return
}
