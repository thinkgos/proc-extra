package cert

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"golang.org/x/crypto/pkcs12"
)

var (
	ErrNotPemEncodedKey = errors.New("cert: Key must be PEM encoded PKCS1 or PKCS8 private key")
	ErrNotRSAPrivateKey = errors.New("cert: Key is not a valid RSA private key")
	ErrNotRSAPublicKey  = errors.New("cert: Key is not a valid RSA public key")
	ErrNotRSAPfxData    = errors.New("cert: pfx data not a valid data")
)

// ParseRSAPrivateKeyFromPEM PEM encoded PKCS1 or PKCS8 private key
// if password exist,PEM encoded PKCS1 or PKCS8 private key protected with password,
// it will decode with password
func ParseRSAPrivateKeyFromPEM(key []byte, password ...string) (*rsa.PrivateKey, error) {
	var err error
	var blockBytes []byte

	block, _ := pem.Decode(key)
	if block == nil {
		return nil, ErrNotPemEncodedKey
	}

	blockBytes = block.Bytes
	if len(password) > 0 {
		//nolint: staticcheck
		blockBytes, err = x509.DecryptPEMBlock(block, []byte(password[0]))
		if err != nil {
			return nil, err
		}
	}

	var parsedKey any
	if parsedKey, err = x509.ParsePKCS1PrivateKey(blockBytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(blockBytes); err != nil {
			return nil, err
		}
	}

	pkey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrNotRSAPrivateKey
	}
	return pkey, nil
}

// ParseRSAPKCS1PrivateKeyFromPEM PEM encoded PKCS1 private key
// if password exist,PEM encoded PKCS1 private key protected with password,
// it will decode with password
func ParseRSAPKCS1PrivateKeyFromPEM(key []byte, password ...string) (*rsa.PrivateKey, error) {
	var err error
	var blockBytes []byte

	block, _ := pem.Decode(key)
	if block == nil {
		return nil, ErrNotPemEncodedKey
	}

	blockBytes = block.Bytes
	if len(password) > 0 {
		//nolint: staticcheck
		blockBytes, err = x509.DecryptPEMBlock(block, []byte(password[0]))
		if err != nil {
			return nil, err
		}
	}
	return x509.ParsePKCS1PrivateKey(blockBytes)
}

// ParseRSAPKCS8PrivateKeyFromPEM PEM encoded PKCS8 private key
// if password exist,PEM encoded PKCS8 private key protected with password,
// it will decode with password
func ParseRSAPKCS8PrivateKeyFromPEM(key []byte, password ...string) (*rsa.PrivateKey, error) {
	var err error
	var blockBytes []byte

	block, _ := pem.Decode(key)
	if block == nil {
		return nil, ErrNotPemEncodedKey
	}

	blockBytes = block.Bytes
	if len(password) > 0 {
		//nolint: staticcheck
		blockBytes, err = x509.DecryptPEMBlock(block, []byte(password[0]))
		if err != nil {
			return nil, err
		}
	}
	parsedKey, err := x509.ParsePKCS8PrivateKey(blockBytes)
	if err != nil {
		return nil, err
	}
	pkey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrNotRSAPrivateKey
	}
	return pkey, nil
}

func ParsePfx(pfxData []byte, password string) (*rsa.PrivateKey, *x509.Certificate, error) {
	pkey, cert, err := pkcs12.Decode(pfxData, password)
	if err != nil {
		return nil, nil, err
	}

	private, ok := pkey.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, ErrNotRSAPfxData
	}
	return private, cert, nil
}

// ParseRSAPublicKeyFromPEM parse public key
// Pem form PKCS1 or PKCS8 public key
func ParseRSAPublicKeyFromPEM(key []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, ErrNotPemEncodedKey
	}
	return parseRSAPublicKey(block.Bytes)
}

// ParseRSAPublicKeyFromDer parse public key
// PKIX, ASN.1 DER form public key
func ParseRSAPublicKeyFromDer(key []byte) (*rsa.PublicKey, error) {
	return parseRSAPublicKey(key)
}

// ParseRSAPublicKey parse public key
// Pem form PKCS1 or PKCS8 public key
// PKIX, ASN.1 DER form public key
func ParseRSAPublicKey(key []byte) (*rsa.PublicKey, error) {
	derBytes := key
	// test if pem form
	block, _ := pem.Decode(key)
	if block != nil {
		derBytes = block.Bytes
	}
	return parseRSAPublicKey(derBytes)
}

// parseRSAPublicKey
// Pem form PKCS1 or PKCS8 public key
// PKIX, ASN.1 DER form public key
func parseRSAPublicKey(derBytes []byte) (*rsa.PublicKey, error) {
	parsedKey, err := x509.ParsePKIXPublicKey(derBytes)
	if err != nil {
		cert, err := x509.ParseCertificate(derBytes)
		if err != nil {
			return nil, err
		}
		parsedKey = cert.PublicKey
	}
	pkey, ok := parsedKey.(*rsa.PublicKey)
	if !ok {
		return nil, ErrNotRSAPublicKey
	}
	return pkey, nil
}
