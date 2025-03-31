package cert

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

func LoadRSAPrivateKeyFromFile(name string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return ParseRSAPrivateKeyFromPEM(keyData)
}

func LoadRSAPublicKeyFromFile(name string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return ParseRSAPublicKey(keyData)
}

func LoadRSAPublicKeyFromPemFile(name string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return ParseRSAPublicKeyFromPEM(keyData)
}

func LoadPfxFromFile(name, password string) (*rsa.PrivateKey, *x509.Certificate, error) {
	keyData, err := os.ReadFile(name)
	if err != nil {
		return nil, nil, err
	}
	return ParsePfx(keyData, password)
}

// 根据文件名解析出证书
// openssl pkcs12 -in xxxx.pfx -clcerts -nokeys -out key.cert
func LoadCertificateFromFile(path string) (*x509.Certificate, error) {
	// Read the verify sign certification key
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Extract the PEM-encoded data block
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("bad key data: not PEM-encoded")
	}
	if got, want := block.Type, "CERTIFICATE"; got != want {
		return nil, fmt.Errorf("unknown key type %q, want %q", got, want)
	}
	// Decode the certification
	return x509.ParseCertificate(block.Bytes)
}
