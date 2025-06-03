// Package sender предоставляет функциональность для отправки метрик на удаленный сервер.
package sender

import (
	"context"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/dvkhr/metrix.git/internal/crypto"
	"github.com/dvkhr/metrix.git/internal/gzip"
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/sign"
	"github.com/dvkhr/metrix.git/internal/storage"
)

// SendOptions содержит параметры для отправки метрик.
type SendOptions struct {
	MemStorage storage.MemStorage
	SignKey    []byte
	PublicKey  *rsa.PublicKey
}

type Strategy interface {
	Send(ctx context.Context, compressedData []byte, signature string) error
}

func Facade(ctx context.Context, sender Strategy, options SendOptions) error {
	logging.Logg.Info("+++Send metrics to server+++")

	allMetrics, err := options.MemStorage.ListSlice(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve metrics: %v", err)
	}
	if len(allMetrics) == 0 {
		logging.Logg.Warn("No metrics available in storage")
		return nil
	}

	jsonMetric, err := json.Marshal(allMetrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	var encryptedData []byte
	if options.PublicKey != nil {
		encryptedStr, err := crypto.EncryptData(jsonMetric, options.PublicKey)
		if err != nil {
			logging.Logg.Error("Failed to encrypt data: %v", err)
			return fmt.Errorf("encryption failed: %w", err)
		}

		encryptedData, err = hex.DecodeString(encryptedStr)
		if err != nil {
			logging.Logg.Error("Failed to decode encrypted data from Hex: %v", err)
			return fmt.Errorf("failed to decode encrypted data: %w", err)
		}
	} else {
		encryptedData = jsonMetric
	}

	compressedData, err := gzip.CompressData(encryptedData)
	if err != nil {
		return fmt.Errorf("compression failed: %w", err)
	}

	signature := sign.SignData(jsonMetric, options.SignKey)

	logging.Logg.Debug("Sending compressed data: %x", compressedData)
	logging.Logg.Debug("Signature: %s", signature)

	options.MemStorage.NewStorage()
	return sender.Send(ctx, compressedData, signature)

}
