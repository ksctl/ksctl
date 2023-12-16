// NOTE: this go is refering to https://shaneutt.com/blog/golang-ca-and-signed-cert-go/
package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

func WriteToFile(buffer *bytes.Buffer, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	n, err := buffer.WriteTo(file)
	if err != nil {
		return err
	}
	fmt.Println("Written bytes=", n)
	return nil
}

func main() {
	privateIPArgs := os.Args[1:]

	var validIPAddresses []net.IP = []net.IP{net.IPv4(127, 0, 0, 1)}
	for _, ip := range privateIPArgs {
		if val := net.ParseIP(string(ip)); val != nil {
			validIPAddresses = append(validIPAddresses, val)
		} else {
			panic("invalid ip address")
		}
	}
	fmt.Printf("%s\n", validIPAddresses)
	fmt.Printf("%#v\n", validIPAddresses)

	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			CommonName: "etcd cluster",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(2, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		panic(err)
	}

	caPEM := new(bytes.Buffer)
	if err := pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}); err != nil {
		panic(err)
	}

	if err := WriteToFile(caPEM, "ca.pem"); err != nil {
		panic(err)
	}

	caPrivKeyPEM := new(bytes.Buffer)
	if err := pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	}); err != nil {
		panic(err)
	}

	// ca.pem and ca-key.pem done
	////////////////////////////////////////////////////////////////////////////
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
		panic(err)
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		panic(err)
	}
	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		panic(err)
	}

	certPrivKeyPEM := new(bytes.Buffer)
	if err := pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	}); err != nil {
		panic(err)
	}
	if err := WriteToFile(certPEM, "etcd.pem"); err != nil {
		panic(err)
	}
	if err := WriteToFile(certPrivKeyPEM, "etcd-key.pem"); err != nil {
		panic(err)
	}
}
