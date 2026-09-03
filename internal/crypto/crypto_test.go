package crypto_test

import (
	"bytes"
	"testing"

	"github.com/GioTld/aldea/internal/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveKey(t *testing.T) {
	passphrase := []byte("test-network-secret")
	salt, err := crypto.NewSalt()
	require.NoError(t, err)

	t.Run("deterministic with same inputs", func(t *testing.T) {
		k1, err := crypto.DeriveKey(passphrase, salt)
		require.NoError(t, err)
		k2, err := crypto.DeriveKey(passphrase, salt)
		require.NoError(t, err)
		assert.Equal(t, k1, k2)
	})

	t.Run("produces 32-byte key", func(t *testing.T) {
		k, err := crypto.DeriveKey(passphrase, salt)
		require.NoError(t, err)
		assert.Len(t, k, 32)
	})

	t.Run("different salt produces different key", func(t *testing.T) {
		salt2, err := crypto.NewSalt()
		require.NoError(t, err)

		k1, err := crypto.DeriveKey(passphrase, salt)
		require.NoError(t, err)
		k2, err := crypto.DeriveKey(passphrase, salt2)
		require.NoError(t, err)
		assert.NotEqual(t, k1, k2)
	})

	t.Run("different passphrase produces different key", func(t *testing.T) {
		k1, err := crypto.DeriveKey(passphrase, salt)
		require.NoError(t, err)
		k2, err := crypto.DeriveKey([]byte("other-secret"), salt)
		require.NoError(t, err)
		assert.NotEqual(t, k1, k2)
	})

	t.Run("rejects short salt", func(t *testing.T) {
		_, err := crypto.DeriveKey(passphrase, []byte("short"))
		assert.ErrorIs(t, err, crypto.ErrInvalidSalt)
	})
}

func TestEncryptDecrypt(t *testing.T) {
	salt, err := crypto.NewSalt()
	require.NoError(t, err)
	key, err := crypto.DeriveKey([]byte("network-secret"), salt)
	require.NoError(t, err)

	plaintext := []byte("hello, Aldea — shard content here")

	t.Run("round trip", func(t *testing.T) {
		ct, err := crypto.Encrypt(plaintext, key)
		require.NoError(t, err)

		got, err := crypto.Decrypt(ct, key)
		require.NoError(t, err)
		assert.Equal(t, plaintext, got)
	})

	t.Run("each call produces unique ciphertext", func(t *testing.T) {
		ct1, err := crypto.Encrypt(plaintext, key)
		require.NoError(t, err)
		ct2, err := crypto.Encrypt(plaintext, key)
		require.NoError(t, err)
		assert.False(t, bytes.Equal(ct1, ct2))
	})

	t.Run("tampered ciphertext is rejected", func(t *testing.T) {
		ct, err := crypto.Encrypt(plaintext, key)
		require.NoError(t, err)
		ct[len(ct)-1] ^= 0xff
		_, err = crypto.Decrypt(ct, key)
		assert.ErrorIs(t, err, crypto.ErrDecryptionFailed)
	})

	t.Run("wrong key is rejected", func(t *testing.T) {
		ct, err := crypto.Encrypt(plaintext, key)
		require.NoError(t, err)

		wrongKey := make([]byte, 32)
		wrongKey[0] = 0xff
		_, err = crypto.Decrypt(ct, wrongKey)
		assert.ErrorIs(t, err, crypto.ErrDecryptionFailed)
	})

	t.Run("ciphertext too short is rejected", func(t *testing.T) {
		_, err := crypto.Decrypt([]byte("too short"), key)
		assert.ErrorIs(t, err, crypto.ErrInvalidCiphertext)
	})

	t.Run("empty plaintext round trip", func(t *testing.T) {
		ct, err := crypto.Encrypt([]byte{}, key)
		require.NoError(t, err)

		got, err := crypto.Decrypt(ct, key)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}
