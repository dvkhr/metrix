package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// EncryptData шифрует данные с использованием публичного ключа.
func EncryptData(data []byte, publicKey *rsa.PublicKey) (string, error) {
	if publicKey == nil {
		return "", fmt.Errorf("public key is nil")
	}

	// Шифрование данных с использованием OAEP
	ciphertext, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		data,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	// Кодирование зашифрованных данных в Hex
	return hex.EncodeToString(ciphertext), nil
}

// DecryptData расшифровывает данные с использованием приватного ключа.
func DecryptData(encryptedData string, privateKey *rsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return nil, errors.New("private key is nil")
	}

	// Декодирование Hex
	ciphertext, err := hex.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex data: %w", err)
	}

	// Расшифровка данных
	plaintext, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privateKey,
		ciphertext,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}
