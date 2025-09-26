package reflux

// EncryptCodec encode a value use PublicKey.
func (r *Reflux) EncryptCodec(v any) (string, error) {
	plainText, err := r.codec.Marshal(v)
	if err != nil {
		return "", err
	}
	return r.Encrypt(plainText)
}

// DecryptCodec decodes to a value use PrivateKey.
func (r *Reflux) DecryptCodec(tk string, v any) error {
	plainText, err := r.Decrypt(tk)
	if err != nil {
		return err
	}
	return r.codec.Unmarshal(plainText, v)
}

// SignCodec sign a message use PrivateKey.
func (r *Reflux) SignCodec(v any) (string, error) {
	plainText, err := r.codec.Marshal(v)
	if err != nil {
		return "", err
	}
	return r.Sign(plainText)
}

// VerifyCodec a message signature use PubicKey.
func (r *Reflux) VerifyCodec(tk string, v any) error {
	plainText, err := r.codec.Marshal(v)
	if err != nil {
		return err
	}
	return r.Verify(tk, plainText)
}
