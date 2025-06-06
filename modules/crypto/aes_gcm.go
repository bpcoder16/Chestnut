package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

func EncryptGCM(plaintext, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 创建 GCM 模式实例
	aesgcm, errGCM := cipher.NewGCM(block)
	if errGCM != nil {
		return "", errGCM
	}

	// 生成随机 nonce（长度由 GCM 要求）
	nonce := make([]byte, aesgcm.NonceSize())
	if _, errR := io.ReadFull(rand.Reader, nonce); errR != nil {
		return "", errR
	}

	// 加密并附加认证标签
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	// 返回 nonce + ciphertext 一起 base64 编码
	final := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(final), nil
}

func DecryptGCM(cipherTextBase64 string, key []byte) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(cipherTextBase64)
	if err != nil {
		return "", err
	}

	block, errC := aes.NewCipher(key)
	if errC != nil {
		return "", errC
	}

	aesgcm, errGCM := cipher.NewGCM(block)
	if errGCM != nil {
		return "", errGCM
	}

	nonceSize := aesgcm.NonceSize()
	if len(raw) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce := raw[:nonceSize]
	cipherText := raw[nonceSize:]

	// 解密并校验标签
	plaintext, errO := aesgcm.Open(nil, nonce, cipherText, nil)
	if errO != nil {
		return "", errO
	}

	return string(plaintext), nil
}
