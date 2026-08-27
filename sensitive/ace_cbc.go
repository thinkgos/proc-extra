package sensitive

import (
	"encoding/json"
	"time"
)

var (
	privacyRandom    = New()
	privacyTimestamp = New(WithIvGen(IvGenTimestamp), WithIvChecker(IvCheckerTimestamp(time.Minute*5)))
)

func EncryptViaRandom(secret, rawText []byte) (string, error) {
	return privacyRandom.Encrypt(secret, rawText)
}

func DecryptViaRandom(secret []byte, cipherText string) ([]byte, error) {
	return privacyRandom.Decrypt(secret, cipherText)
}

func EncryptObjectViaRandom(secret []byte, data any) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return EncryptViaRandom(secret, b)
}

func DecryptObjectViaRandom(secret []byte, cipherText string, v any) error {
	bs, err := DecryptViaRandom(secret, cipherText)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, v)
}

func EncryptViaTimestamp(secret, rawText []byte) (string, error) {
	return privacyTimestamp.Encrypt(secret, rawText)
}

func DecryptViaTimestamp(secret []byte, cipherText string) ([]byte, error) {
	return privacyTimestamp.Decrypt(secret, cipherText)
}

func EncryptObjectViaTimestamp(secret []byte, data any) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return EncryptViaTimestamp(secret, b)
}

func DecryptObjectViaTimestamp(secret []byte, cipherText string, v any) error {
	b, err := DecryptViaTimestamp(secret, cipherText)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
