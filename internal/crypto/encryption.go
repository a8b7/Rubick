package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"sync"
)

var (
	encryptionKey []byte
	once          sync.Once
	keyErr        error
)

// InitKey 初始化加密密钥
func InitKey() error {
	var initErr error
	once.Do(func() {
		// 从环境变量获取密钥
		key := os.Getenv("RUBICK_ENCRYPTION_KEY")
		if key == "" {
			initErr = errors.New("RUBICK_ENCRYPTION_KEY environment variable is required")
			return
		}

		// 密钥必须是 16, 24 或 32 字节
		if len(key) < 32 {
			// 填充到 32 字节
			padded := make([]byte, 32)
			copy(padded, key)
			encryptionKey = padded
		} else {
			encryptionKey = []byte(key[:32])
		}
	})

	return initErr
}

// Encrypt 加密字符串
func Encrypt(plaintext string) (string, error) {
	if err := InitKey(); err != nil {
		return "", err
	}

	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密字符串
func Decrypt(ciphertext string) (string, error) {
	if err := InitKey(); err != nil {
		return "", err
	}

	if ciphertext == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("密文太短")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
