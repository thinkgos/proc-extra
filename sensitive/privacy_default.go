package sensitive

import "time"

var (
	privacyRandom    = New()
	privacyTimestamp = New(WithIvGen(IvGenTimestamp), WithIvChecker(IvCheckerTimestamp(time.Minute*5)))
)

func EncryptPrivacyRandom(secret, rawText []byte) (string, error) {
	return privacyRandom.Encrypt(secret, rawText)
}

func DecryptPrivacyRandom(secret []byte, cipherText string) ([]byte, error) {
	return privacyRandom.Decrypt(secret, cipherText)
}

func EncryptPrivacyTimestamp(secret, rawText []byte) (string, error) {
	return privacyTimestamp.Encrypt(secret, rawText)
}

func DecryptPrivacyTimestamp(secret []byte, cipherText string) ([]byte, error) {
	return privacyTimestamp.Decrypt(secret, cipherText)
}
