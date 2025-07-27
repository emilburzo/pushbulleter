package pushbullet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

type E2EManager struct {
	key []byte
}

func NewE2EManager(password string) *E2EManager {
	return NewE2EManagerWithSalt(password, "pushbullet")
}

func NewE2EManagerWithSalt(password, salt string) *E2EManager {
	// Use PBKDF2 with the provided salt (should be user iden in production)
	key := pbkdf2.Key([]byte(password), []byte(salt), 30000, 32, sha256.New)
	
	return &E2EManager{
		key: key,
	}
}

func (e *E2EManager) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	if len(data) < 29 { // 1 (version) + 16 (tag) + 12 (iv) minimum
		return "", fmt.Errorf("ciphertext too short")
	}

	// Check version
	if data[0] != '1' {
		return "", fmt.Errorf("unsupported encryption version")
	}

	tag := data[1:17]
	iv := data[17:29]
	encrypted := data[29:]

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// For GCM, the tag is appended to the ciphertext
	ciphertextWithTag := append(encrypted, tag...)

	plaintext, err := gcm.Open(nil, iv, ciphertextWithTag, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

func (e *E2EManager) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	iv := make([]byte, 12)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	ciphertext := gcm.Seal(nil, iv, []byte(plaintext), nil)
	
	// Split ciphertext and tag
	tagStart := len(ciphertext) - 16
	encrypted := ciphertext[:tagStart]
	tag := ciphertext[tagStart:]

	// Format: version + tag + iv + encrypted
	result := make([]byte, 1+16+12+len(encrypted))
	result[0] = '1'
	copy(result[1:17], tag)
	copy(result[17:29], iv)
	copy(result[29:], encrypted)

	return base64.StdEncoding.EncodeToString(result), nil
}
