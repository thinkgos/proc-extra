package excel

import (
	"crypto/rand"
	"time"
	"unsafe"
)

var defaultAlphabet = []byte("QWERTYUIOPLKJHGFDSAZXCVBNMabcdefghijklmnopqrstuvwxyz")

func randAlphabet(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err == nil {
		for i, v := range b {
			b[i] = defaultAlphabet[v%byte(len(defaultAlphabet))]
		}
	}
	return *(*string)(unsafe.Pointer(&b))
}

func randExcelFilename() string {
	return "test_" + randAlphabet(10) + ".xlsx"
}

func customTitle() (string, error) {
	return time.Now().Format(time.DateTime), nil
}
