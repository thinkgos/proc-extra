package reflux

import (
	"google.golang.org/protobuf/proto"
)

// EncryptProto encode a protobuf message use PublicKey.
func (r *Reflux) EncryptProto(v proto.Message) (string, error) {
	plainText, err := proto.Marshal(v)
	if err != nil {
		return "", err
	}
	return r.Encrypt(plainText)
}

// DecryptProto decodes to a protobuf message use PrivateKey.
func (r *Reflux) DecryptProto(tk string, v proto.Message) error {
	plainText, err := r.Decrypt(tk)
	if err != nil {
		return err
	}
	return proto.Unmarshal(plainText, v)
}

// SignProto sign a protobuf message use PrivateKey.
func (r *Reflux) SignProto(v proto.Message) (string, error) {
	plainText, err := proto.Marshal(v)
	if err != nil {
		return "", err
	}
	return r.Sign(plainText)
}

// VerifyProto a protobuf message signature use PublicKey.
func (r *Reflux) VerifyProto(tk string, v proto.Message) error {
	plainText, err := proto.Marshal(v)
	if err != nil {
		return err
	}
	return r.Verify(tk, plainText)
}
