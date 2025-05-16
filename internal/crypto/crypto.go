package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// ReadPublicKey читает и возвращает публичный ключ из файла.
func ReadPublicKey(filePath string) (*rsa.PublicKey, error) {
	// Чтение файла с публичным ключом
	keyBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	// Декодирование PEM-блока
	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	// Парсинг публичного ключа
	pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return pubKey, nil
}

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

// ReadPrivateKey читает и возвращает приватный ключ из файла.
func ReadPrivateKey(filePath string) (*rsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return privKey, nil
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
