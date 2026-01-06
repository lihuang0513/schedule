package tool

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

// Aes/ECB模式的加密方法，PKCS7填充方式
func AesEncrypt(src, key []byte) ([]byte, error) {
	Block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(src) == 0 {
		return nil, errors.New("plaintext empty")
	}
	mode := NewECBEncrypter(Block)
	ciphertext := src
	mode.CryptBlocks(ciphertext, ciphertext)
	return ciphertext, nil
}

// Aes/ECB模式的解密方法，PKCS7填充方式
func AesDecrypt(src, key []byte) ([]byte, error) {
	Block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(src) == 0 {
		return nil, errors.New("plaintext empty")
	}
	mode := NewECBDecrypter(Block)
	ciphertext := src
	mode.CryptBlocks(ciphertext, ciphertext)
	return ciphertext, nil
}

// AesDecrypt2 用于cookie解密
func AesDecrypt2(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(src) == 0 {
		return nil, errors.New("ciphertext is empty")
	}
	decrypted := make([]byte, len(src))
	mode := NewECBDecrypter(block)
	mode.CryptBlocks(decrypted, src)
	return PKCS7UnPadding2(decrypted)
}

// PKCS7UnPadding2 removes PKCS7 padding from the plaintext.（用于cookie解密）
func PKCS7UnPadding2(ciphertext []byte) ([]byte, error) {
	length := len(ciphertext)
	unpadding := int(ciphertext[length-1])
	if unpadding > length {
		return nil, errors.New("unpadding size is larger than data length")
	}
	return ciphertext[:(length - unpadding)], nil
}

// RepeatKeyToBytes extends the key to the desired length by repeating or truncating the input key.
func RepeatKeyToBytes(key []byte, length int) []byte {
	if len(key) >= length {
		return key[:length]
	}
	padding := make([]byte, length-len(key))
	return append(key, padding...)
}

func DecryptAES(encodedStr, secret string, keySize int) (string, error) {
	key := []byte(secret)

	if keySize == 256 {
		key = RepeatKeyToBytes(key, 32) // Expand key to 32 bytes for AES-256
	} else if keySize == 128 {
		key = RepeatKeyToBytes(key, 16) // Ensure key is 16 bytes for AES-128
	} else {
		return "", errors.New("unsupported key size, only 128 or 256 are allowed")
	}

	cipherText, err := base64.StdEncoding.DecodeString(encodedStr)
	if err != nil {
		return "", err
	}

	decryptedData, err := AesDecrypt2(cipherText, key)
	if err != nil {
		return "", err
	}

	return string(decryptedData), nil
}

// ECB模式结构体
type ecb struct {
	b         cipher.Block
	blockSize int
}

// 实例化ECB对象
func newECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

// ECB加密类
type ecbEncrypter ecb

func NewECBEncrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbEncrypter)(newECB(b))
}

func (x *ecbEncrypter) BlockSize() int {
	return x.blockSize
}

func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.blockSize])
		dst = dst[x.blockSize:]
		src = src[x.blockSize:]
	}
}

// ECB解密类
type ecbDecrypter ecb

func NewECBDecrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbDecrypter)(newECB(b))
}

func (x *ecbDecrypter) BlockSize() int {
	return x.blockSize
}

func (x *ecbDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for len(src) > 0 {
		x.b.Decrypt(dst, src[:x.blockSize])
		dst = dst[x.blockSize:]
		src = src[x.blockSize:]
	}
}

// PKCS7填充
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS7去除
func PKCS7UnPadding(ciphertext []byte) []byte {
	length := len(ciphertext)
	unpadding := int(ciphertext[length-1])
	return ciphertext[:(length - unpadding)]
}

// 零点填充
func ZerosPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(0)}, padding)
	return append(ciphertext, padtext...)
}

// 零点去除
func ZerosUnPadding(ciphertext []byte) []byte {
	return bytes.TrimFunc(ciphertext, func(r rune) bool {
		return r == rune(0)
	})
}
