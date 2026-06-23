package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

// CryptoService provides AES-256-GCM encryption/decryption for account credentials.
// The key is derived from the provided key material via SHA-256, giving a consistent
// 32-byte key suitable for AES-256 regardless of input length.
type CryptoService struct {
	key []byte
}

// New creates a CryptoService by hashing the key material with SHA-256 to produce
// a 32-byte AES-256 key.
func New(keyMaterial string) *CryptoService {
	h := sha256.Sum256([]byte(keyMaterial))
	return &CryptoService{key: h[:]}
}

// NewFromPassword derives a 32-byte key from a password and salt using Argon2id.
// time=3, memory=64MB, threads=4 are the parameters; adjust for your threat model.
func NewFromPassword(password, salt []byte) *CryptoService {
	key := argon2.IDKey(password, salt, 3, 64*1024, 4, 32)
	return &CryptoService{key: key}
}

// Encrypt encrypts plaintext with AES-256-GCM. The returned ciphertext is
// nonce || encrypted_data || auth_tag. A fresh random nonce is generated for
// every call, so identical plaintexts produce different ciphertexts.
func (s *CryptoService) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal appends the encrypted data to the nonce: nonce || ciphertext || tag
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts ciphertext produced by Encrypt. It expects the format
// nonce || encrypted_data || auth_tag. Returns an error if the ciphertext
// is too short or if authentication fails (tampered data).
func (s *CryptoService) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, cipherData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, cipherData, nil)
}

// EncryptString encrypts a plaintext string and returns a base64-encoded
// representation of the ciphertext (nonce + encrypted data + tag).
func (s *CryptoService) EncryptString(plaintext string) (string, error) {
	b, err := s.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// DecryptString decodes a base64 ciphertext string and decrypts it.
func (s *CryptoService) DecryptString(encoded string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	decrypted, err := s.Decrypt(b)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}
