package signature

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
)

// Hmac Generate a hex hash value with the key,
// expect: hmacmd5, hmacsha1, hmacsha224, hmacsha256, hmacsha384, hmacsha512.
// default: hmacsha256
func Hmac(method string, key, planText []byte) []byte {
	var h func() hash.Hash

	switch method {
	case "hmacmd5":
		h = md5.New
	case "hmacsha1":
		h = sha1.New
	case "hmacsha224":
		h = sha256.New224
	case "hmacsha256":
		h = sha256.New
	case "hmacsha384":
		h = sha512.New384
	case "hmacsha512":
		h = sha512.New
	default:
		h = sha512.New
	}
	hasher := hmac.New(h, key)
	_, _ = hasher.Write(planText)
	return hasher.Sum(nil)
}
