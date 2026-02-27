package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"lingosql/internal/config"
)

// Encrypt 使用 AES 加密数据
func Encrypt(plaintext string) (string, error) {
	cfg := config.GetConfig()
	key := []byte(cfg.Encryption.Key)
	
	// 确保密钥长度为 32 字节（AES-256）
	if len(key) != 32 {
		return "", errors.New("加密密钥长度必须为 32 字节")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 创建 GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 创建随机 nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密数据
func Decrypt(ciphertext string) (string, error) {
	cfg := config.GetConfig()
	key := []byte(cfg.Encryption.Key)
	
	// 确保密钥长度为 32 字节
	if len(key) != 32 {
		return "", errors.New("加密密钥长度必须为 32 字节")
	}

	// 解码 base64
	encrypted, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(encrypted) < nonceSize {
		return "", errors.New("密文长度无效")
	}

	nonce, ciphertextBytes := encrypted[:nonceSize], encrypted[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
