package reflux

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"

	"github.com/thinkgos/proc-extra/cert"
)

type CodecString interface {
	EncodeToString([]byte) string
	DecodeString(string) ([]byte, error)
}

type Codec interface {
	Marshal(any) ([]byte, error)
	Unmarshal([]byte, any) error
}

type Option func(*Reflux)

// WithCodecString set the codec string. default use base64.StdEncoding.
func WithCodecString(c CodecString) Option {
	return func(r *Reflux) {
		if c != nil {
			r.codecString = c
		}
	}
}

// WithCodec set the codec. default use CodecJSON.
func WithCodec(c Codec) Option {
	return func(r *Reflux) {
		if c != nil {
			r.codec = c
		}
	}
}

type Reflux struct {
	priv        *rsa.PrivateKey
	pub         *rsa.PublicKey
	codec       Codec
	codecString CodecString
}

// New returns a new Reflux.
// privKey, pubKey: string or filepath string.
func New(privKey, pubKey string, opts ...Option) (*Reflux, error) {
	priv, err := cert.ParseRSAPrivateKeyFromPEM([]byte(privKey))
	if err != nil {
		priv, err = cert.LoadRSAPrivateKeyFromFile(privKey)
		if err != nil {
			return nil, err
		}
	}
	pub, err := cert.ParseRSAPublicKeyFromPEM([]byte(pubKey))
	if err != nil {
		pub, err = cert.LoadRSAPublicKeyFromPemFile(pubKey)
		if err != nil {
			return nil, err
		}
	}
	r := &Reflux{
		priv:        priv,
		pub:         pub,
		codec:       CodecJSON{},
		codecString: base64.StdEncoding,
	}
	for _, f := range opts {
		f(r)
	}
	return r, nil
}

func (r *Reflux) PrivateKey() *rsa.PrivateKey { return r.priv }

func (r *Reflux) PublicKey() *rsa.PublicKey { return r.pub }

// Encrypt encode a value use PublicKey.
func (r *Reflux) Encrypt(plainText []byte) (string, error) {
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, r.pub, plainText)
	if err != nil {
		return "", err
	}
	return r.codecString.EncodeToString(cipherText), nil
}

// Decrypt decodes to a value use PrivateKey.
func (r *Reflux) Decrypt(tk string) ([]byte, error) {
	cipherText, err := r.codecString.DecodeString(tk)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, r.priv, cipherText)
}

// Sign sign a message use PrivateKey.
func (r *Reflux) Sign(plainText []byte) (string, error) {
	hashed := sha256.Sum256(plainText)
	sighText, err := rsa.SignPKCS1v15(rand.Reader, r.priv, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return r.codecString.EncodeToString(sighText), nil
}

// Verify a message signature use PubicKey.
func (r *Reflux) Verify(tk string, plainText []byte) error {
	signText, err := r.codecString.DecodeString(tk)
	if err != nil {
		return err
	}
	hashed := sha256.Sum256(plainText)
	return rsa.VerifyPKCS1v15(r.pub, crypto.SHA256, hashed[:], signText)
}

// Encrypt rsa PKCS #1 v1.5. and base64 encoded.
func Encrypt(pub *rsa.PublicKey, rawText []byte) (string, error) {
	b, err := rsa.EncryptPKCS1v15(rand.Reader, pub, rawText)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// Decrypt base64 decoded and rsa PKCS #1 v1.5.
func Decrypt(pri *rsa.PrivateKey, cipherText string) ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, pri, b)
}
