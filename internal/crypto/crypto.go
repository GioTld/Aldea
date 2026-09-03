package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

var (
	ErrInvalidCiphertext = errors.New("ciphertext too short or malformed")
	ErrDecryptionFailed  = errors.New("decryption failed: authentication tag mismatch")
	ErrInvalidSalt       = errors.New("salt must be at least 16 bytes")
)

// Argon2id parameters chosen to resist offline brute-force on consumer hardware.
// Reviewed in docs/adr/0001-symmetric-cipher-selection.md (RNF-02).
const (
	argon2Time    uint32 = 3
	argon2Memory  uint32 = 64 * 1024
	argon2Threads uint8  = 4
	keySize       uint32 = 32
	minSaltSize          = 16
)

func NewSalt() ([]byte, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generating salt: %w", err)
	}
	return salt, nil
}

func DeriveKey(passphrase, salt []byte) ([]byte, error) {
	if len(salt) < minSaltSize {
		return nil, ErrInvalidSalt
	}
	return argon2.IDKey(passphrase, salt, argon2Time, argon2Memory, argon2Threads, keySize), nil
}

func Encrypt(plaintext, key []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generating nonce: %w", err)
	}

	return aead.Seal(nonce, nonce, plaintext, nil), nil
}

func Decrypt(ciphertext, key []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	if len(ciphertext) < aead.NonceSize() {
		return nil, ErrInvalidCiphertext
	}

	nonce, ct := ciphertext[:aead.NonceSize()], ciphertext[aead.NonceSize():]
	plaintext, err := aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}
