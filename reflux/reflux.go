package reflux

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"

	"google.golang.org/protobuf/proto"

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

func WithCodecString(c CodecString) Option {
	return func(r *Reflux) {
		if c != nil {
			r.codecString = c
		}
	}
}

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

// Encrypt encode a message use PublicKey.
func (r *Reflux) Encrypt(message proto.Message) (string, error) {
	plainText, err := r.codec.Marshal(message)
	if err != nil {
		return "", err
	}
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, r.pub, plainText)
	if err != nil {
		return "", err
	}
	return r.codecString.EncodeToString(cipherText), nil
}

// Decrypt decodes to a message use PrivateKey.
func (r *Reflux) Decrypt(tk string, message proto.Message) error {
	cipherText, err := r.codecString.DecodeString(tk)
	if err != nil {
		return err
	}
	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, r.priv, cipherText)
	if err != nil {
		return err
	}
	return r.codec.Unmarshal(plainText, message)
}

// Sign sign a message use PrivateKey.
func (r *Reflux) Sign(message proto.Message) (string, error) {
	plainText, err := r.codec.Marshal(message)
	if err != nil {
		return "", err
	}
	hashed := sha256.Sum256(plainText)
	sighText, err := rsa.SignPKCS1v15(rand.Reader, r.priv, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return r.codecString.EncodeToString(sighText), nil
}

// Verify a message signature use PubicKey.
func (r *Reflux) Verify(tk string, message proto.Message) error {
	plainText, err := r.codec.Marshal(message)
	if err != nil {
		return err
	}
	sighText, err := r.codecString.DecodeString(tk)
	if err != nil {
		return err
	}
	hashed := sha256.Sum256(plainText)
	return rsa.VerifyPKCS1v15(r.pub, crypto.SHA256, hashed[:], sighText)
}

// EncryptProto encode a protobuf message use PublicKey.
func (r *Reflux) EncryptProto(message proto.Message) (string, error) {
	plainText, err := proto.Marshal(message)
	if err != nil {
		return "", err
	}
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, r.pub, plainText)
	if err != nil {
		return "", err
	}
	return r.codecString.EncodeToString(cipherText), nil
}

// DecryptProto decodes to a protobuf message use PrivateKey.
func (r *Reflux) DecryptProto(tk string, message proto.Message) error {
	cipherText, err := r.codecString.DecodeString(tk)
	if err != nil {
		return err
	}
	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, r.priv, cipherText)
	if err != nil {
		return err
	}
	return proto.Unmarshal(plainText, message)
}

// SignProto sign a protobuf message use PrivateKey.
func (r *Reflux) SignProto(message proto.Message) (string, error) {
	plainText, err := proto.Marshal(message)
	if err != nil {
		return "", err
	}
	hashed := sha256.Sum256(plainText)
	sighText, err := rsa.SignPKCS1v15(rand.Reader, r.priv, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return r.codecString.EncodeToString(sighText), nil
}

// VerifyProto a protobuf message signature use PublicKey.
func (r *Reflux) VerifyProto(tk string, message proto.Message) error {
	plainText, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	sighText, err := r.codecString.DecodeString(tk)
	if err != nil {
		return err
	}
	hashed := sha256.Sum256(plainText)
	return rsa.VerifyPKCS1v15(r.pub, crypto.SHA256, hashed[:], sighText)
}
