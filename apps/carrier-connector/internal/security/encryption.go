package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// EncryptionService provides data encryption at rest
type EncryptionService struct {
	masterKey []byte
	gcm       cipher.AEAD
}

// EncryptionConfig configures encryption
type EncryptionConfig struct {
	MasterKey string
	KeyDerivationSalt string
}

// NewEncryptionService creates a new encryption service
func NewEncryptionService(config EncryptionConfig) (*EncryptionService, error) {
	salt := []byte(config.KeyDerivationSalt)
	if len(salt) == 0 {
		salt = []byte("telecom-platform-default-salt")
	}

	key := pbkdf2.Key([]byte(config.MasterKey), salt, 100000, 32, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &EncryptionService{
		masterKey: key,
		gcm:       gcm,
	}, nil
}

// Encrypt encrypts plaintext data
func (e *EncryptionService) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := e.gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext data
func (e *EncryptionService) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < e.gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:e.gcm.NonceSize()]
	ciphertext = ciphertext[e.gcm.NonceSize():]

	plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// EncryptString encrypts a string and returns base64
func (e *EncryptionService) EncryptString(plaintext string) (string, error) {
	encrypted, err := e.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptString decrypts a base64 string
func (e *EncryptionService) DecryptString(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	decrypted, err := e.Decrypt(data)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

// FieldEncryptor provides field-level encryption for sensitive data
type FieldEncryptor struct {
	service *EncryptionService
	fields  map[string]bool
}

// NewFieldEncryptor creates a field encryptor
func NewFieldEncryptor(service *EncryptionService, sensitiveFields []string) *FieldEncryptor {
	fields := make(map[string]bool)
	for _, f := range sensitiveFields {
		fields[f] = true
	}
	return &FieldEncryptor{
		service: service,
		fields:  fields,
	}
}

// IsSensitive checks if a field is sensitive
func (fe *FieldEncryptor) IsSensitive(field string) bool {
	return fe.fields[field]
}

// EncryptField encrypts a field value if sensitive
func (fe *FieldEncryptor) EncryptField(field, value string) (string, error) {
	if !fe.IsSensitive(field) {
		return value, nil
	}
	return fe.service.EncryptString(value)
}

// DecryptField decrypts a field value if sensitive
func (fe *FieldEncryptor) DecryptField(field, value string) (string, error) {
	if !fe.IsSensitive(field) {
		return value, nil
	}
	return fe.service.DecryptString(value)
}

// DefaultSensitiveFields returns commonly sensitive fields
func DefaultSensitiveFields() []string {
	return []string{
		"ssn", "social_security_number",
		"credit_card", "card_number", "cvv",
		"password", "secret", "api_key",
		"phone", "email", "address",
		"date_of_birth", "dob",
		"bank_account", "routing_number",
		"iccid", "imsi", "msisdn",
	}
}
