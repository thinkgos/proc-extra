package sensitive

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strconv"
	"time"
)

// error defined
var (
	ErrInputNotMultipleBlocks = errors.New("decoded message length must be multiple of block size")
	ErrIvInvalidSize          = errors.New("iv length must equal block size")
	ErrUnPaddingOutOfRange    = errors.New("unPadding out of range")
	ErrIvValueIllegal         = errors.New("iv value illegal")
	ErrIvValueExpired         = errors.New("iv value has expired")
)

type Privacy struct {
	ivGen     func(blockSize int) ([]byte, error)
	ivChecker func(iv []byte) error
}

type Option func(p *Privacy)

func WithIvGen(f func(blockSize int) ([]byte, error)) Option {
	return func(p *Privacy) {
		if f != nil {
			p.ivGen = f
		}
	}
}

func WithIvChecker(f func(iv []byte) error) Option {
	return func(p *Privacy) {
		if f != nil {
			p.ivChecker = f
		}
	}
}

func New(opts ...Option) *Privacy {
	p := &Privacy{
		ivGen:     IvGenRandom,
		ivChecker: IvCheckerTrue,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Encrypt aes cbc, iv + ciphertext base64 encoded.
// key must 16, 24, 32
func (p *Privacy) Encrypt(secret, rawText []byte) (string, error) {
	cip, err := aes.NewCipher(secret)
	if err != nil {
		return "", err
	}
	blockSize := cip.BlockSize()

	// 生成随机iv
	iv, err := p.ivGen(blockSize)
	if err != nil {
		return "", err
	}
	if len(iv) != blockSize {
		return "", ErrIvInvalidSize
	}
	orig := PCKSPadding(rawText, blockSize)
	cipherText := make([]byte, blockSize+len(orig))
	copy(cipherText[:blockSize], iv)
	cipher.NewCBCEncrypter(cip, iv).CryptBlocks(cipherText[blockSize:], orig)
	// iv + ciphertext 一起进行 base64
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt aes cbc, base64 decoded iv + ciphertext.
// key must 16, 24, 32
func (p *Privacy) Decrypt(secret []byte, cipherText string) ([]byte, error) {
	body, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return nil, err
	}
	cip, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}
	blockSize := cip.BlockSize()
	if len(body) == 0 || len(body)%blockSize != 0 {
		return nil, ErrInputNotMultipleBlocks
	}
	iv, msg := body[:blockSize], body[blockSize:]
	err = p.ivChecker(iv)
	if err != nil {
		return nil, err
	}
	cipher.NewCBCDecrypter(cip, iv).CryptBlocks(msg, msg)
	return PCKSUnPadding(msg, blockSize)
}

// PCKSPadding PKCS#5和PKCS#7 填充
func PCKSPadding(origData []byte, blockSize int) []byte {
	padSize := blockSize - len(origData)%blockSize
	padText := bytes.Repeat([]byte{byte(padSize)}, padSize)
	return append(origData, padText...)
}

// PCKSUnPadding PKCS#5和PKCS#7 解填充
func PCKSUnPadding(origData []byte, blockSize int) ([]byte, error) {
	orgLen := len(origData)
	if orgLen == 0 {
		return nil, ErrUnPaddingOutOfRange
	}
	unPadSize := int(origData[orgLen-1])
	if unPadSize > blockSize || unPadSize > orgLen {
		return nil, ErrUnPaddingOutOfRange
	}
	for _, v := range origData[orgLen-unPadSize:] {
		if v != byte(unPadSize) {
			return nil, ErrUnPaddingOutOfRange
		}
	}
	return origData[:(orgLen - unPadSize)], nil
}

// IvGenRandom 随机生成iv
func IvGenRandom(blockSize int) ([]byte, error) {
	v := make([]byte, blockSize)
	_, err := rand.Read(v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// IvCheckerTrue 校验iv是否合法, 永远成功
func IvCheckerTrue(iv []byte) error { return nil }

// IvGenTimestamp 生成iv, 前三个字节随机, 后13个字节采用毫秒时间戳
func IvGenTimestamp(blockSize int) ([]byte, error) {
	// 前三个字节随机
	rd := make([]byte, 3)
	_, err := rand.Read(rd)
	if err != nil {
		return nil, err
	}
	// 后13个字节采用毫秒时间戳
	s := strconv.FormatInt(time.Now().UnixMilli(), 10)
	return append(rd, []byte(s)...), nil
}

// IvCheckerTimestamp 校验iv是否合法, 校验通过返回nil, 校验失败返回错误
func IvCheckerTimestamp(gap time.Duration) func(iv []byte) error {
	return func(iv []byte) error {
		if len(iv) < 3 {
			return ErrIvValueIllegal
		}
		// 去除前三个随机的字节
		ts := string(iv[3:])
		t, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			return ErrIvValueIllegal
		}
		if t+int64(gap/time.Millisecond) <= time.Now().UnixMilli() {
			return ErrIvValueExpired
		}
		return nil
	}
}
