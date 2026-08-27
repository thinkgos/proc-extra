package sensitive

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// Encrypt aes gcm, base64 encoded.
// secret must 16, 24, 32
func EncryptGCM(secret, plaintext []byte) (string, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// GCM 标准 Nonce 长度为 12 字节 (gcm.NonceSize())
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt aes cbc, base64 decoded iv + ciphertext.
// secret must 16, 24, 32
func DecryptGCM(secret []byte, cipherText string) ([]byte, error) {
	body, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(body) < nonceSize {
		return nil, ErrInputIllegal
	}
	nonce, ciphertext := body[:nonceSize], body[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
