package desensitize

// 定义敏感加密接口, 对象单个字段加密
type SensitiveEncryptAble interface {
	IntoSensitive(identityId string) error
}

// 定义敏感加密接口, 对象单个字段解密
type SensitiveDecryptAble interface {
	FromSensitive(identityId string) error
}

// 定义敏感加密接口, 整个对象序列化后加密
type SensitiveEncryptObjectAble interface {
	EncryptSensitive(identityId string) (string, error)
}

// 定义敏感解密接口, 整个对象解密后反序列化
type SensitiveDecryptObjectAble interface {
	DecryptSensitive(cipherText string, identityId string) error
}
