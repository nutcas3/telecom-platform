package services

import (
	"crypto/aes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
)

// generateAuthKeys generates authentication keys for the subscriber
func (s *SubscriberService) generateAuthKeys() (string, string, error) {
	// Generate 128-bit random key (K)
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return "", "", err
	}

	// Get OP (Operator variant) from operator configuration
	// This should be consistent across all subscribers for the same operator
	op := s.getOperatorVariant()
	if op == nil {
		return "", "", fmt.Errorf("operator variant not configured")
	}

	// Generate OPc (derived from OP and K) using AES-128 encryption
	// OPc = AES-128(K, OP) where OP is encrypted with K
	opc, err := s.generateOPc(key, op)
	if err != nil {
		return "", "", err
	}

	return hex.EncodeToString(key), hex.EncodeToString(opc), nil
}

// getOperatorVariant returns the operator variant (OP) from configuration
func (s *SubscriberService) getOperatorVariant() []byte {
	// Get operator variant from environment variable or secure configuration
	// This should be the same across all subscribers for the same operator
	opStr := os.Getenv("OPERATOR_VARIANT")
	if opStr == "" {
		// Fallback to default for development
		opStr = "TelecomOP1234567" // 16-byte operator variant
	}
	
	// Ensure exactly 16 bytes for AES-128
	op := make([]byte, 16)
	copy(op, []byte(opStr))
	
	// If shorter than 16 bytes, pad with zeros
	if len(opStr) < 16 {
		for i := len(opStr); i < 16; i++ {
			op[i] = 0
		}
	}
	
	return op
}

// generateOPc derives OPc from OP and K using AES-128 encryption
func (s *SubscriberService) generateOPc(k, op []byte) ([]byte, error) {
	// Create AES-128 cipher block with key K
	block, err := aes.NewCipher(k)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// OPc = AES-128(K, OP) - encrypt OP with key K
	opc := make([]byte, aes.BlockSize) // AES block size is 16 bytes
	
	// Create ECB mode cipher (as per 3GPP specification for OPc derivation)
	// In 3GPP, OPc is derived using AES-128 ECB mode
	if len(op) != aes.BlockSize {
		return nil, fmt.Errorf("OP must be 16 bytes for AES-128")
	}

	// Encrypt OP using ECB mode (single block)
	block.Encrypt(opc, op)

	return opc, nil
}
