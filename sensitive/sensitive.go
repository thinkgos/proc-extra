package sensitive

// EncryptSensitiveAble 敏感加密接口
type EncryptSensitiveAble interface {
	EncryptSensitive(secret []byte) (string, error)
}

// DecryptSensitiveAble 敏感解密接口
type DecryptSensitiveAble interface {
	DecryptSensitive(secret []byte, cipherText string) error
}
