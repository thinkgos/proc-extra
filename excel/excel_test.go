package excel

import (
	"crypto/rand"
	"text/template"
	"time"
	"unsafe"
)

var defaultAlphabet = []byte("QWERTYUIOPLKJHGFDSAZXCVBNMabcdefghijklmnopqrstuvwxyz")
var tmpl = template.Must(template.New("customTitle").Parse(`{{.Now}}`))

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

func customTitle() string {
	return time.Now().Format(time.DateTime)
}
