package ssh

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

// generatePrivateKey creates a RSA Private Key of specified byte size
func generatePrivateKey(ctx context.Context, log types.LoggerFactory, bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, ksctlErrors.ErrSSHExec.Wrap(
			log.NewError(ctx, "failed to generate key", "Reason", err),
		)
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, ksctlErrors.ErrSSHExec.Wrap(
			log.NewError(ctx, "failed to validate key", "Reason", err),
		)
	}

	log.Print(ctx, "Private Key helper-gen")
	return privateKey, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePrivateKeyToPEM(log types.LoggerFactory, privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func generatePublicKey(ctx context.Context, log types.LoggerFactory, privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, ksctlErrors.ErrSSHExec.Wrap(
			log.NewError(ctx, "failed to create public key for given private key", "Reason", err),
		)
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	log.Print(ctx, "Public key helper-gen")
	return pubKeyBytes, nil
}

func CreateSSHKeyPair(ctx context.Context, log types.LoggerFactory, state *storageTypes.StorageDocument) error {

	bitSize := 4096

	privateKey, err := generatePrivateKey(ctx, log, bitSize)
	if err != nil {
		return err
	}

	publicKeyBytes, err := generatePublicKey(ctx, log, &privateKey.PublicKey)
	if err != nil {
		return err
	}

	privateKeyBytes := encodePrivateKeyToPEM(log, privateKey)

	log.Debug(ctx, "Printing", "ssh pub key", string(publicKeyBytes))
	log.Debug(ctx, "Printing", "ssh private key", string(privateKeyBytes))

	state.SSHKeyPair.PrivateKey = string(privateKeyBytes)
	state.SSHKeyPair.PublicKey = string(publicKeyBytes)

	return nil
}
