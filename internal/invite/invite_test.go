package invite_test

import (
	"testing"
	"time"

	"github.com/GioTld/aldea/internal/invite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testKey = []byte("aldea-tracker-secret-key-32bytes!")

func TestInviteToken(t *testing.T) {
	t.Run("creates and parses valid invite token", func(t *testing.T) {
		mgr := invite.NewTokenManager(testKey)
		tokStr, err := mgr.CreateToken("127.0.0.1:9000", []byte("network-key-12345"), 1*time.Hour, 1)
		require.NoError(t, err)
		assert.NotEmpty(t, tokStr)

		tok, err := mgr.ValidateAndUse(tokStr)
		require.NoError(t, err)
		assert.Equal(t, "127.0.0.1:9000", tok.TrackerAddr)
		assert.Equal(t, []byte("network-key-12345"), tok.NetworkKey)
	})

	t.Run("expired token returns error", func(t *testing.T) {
		mgr := invite.NewTokenManager(testKey)
		tokStr, err := mgr.CreateToken("127.0.0.1:9000", []byte("network-key-12345"), -1*time.Minute, 1)
		require.NoError(t, err)

		_, err = mgr.ValidateAndUse(tokStr)
		require.ErrorIs(t, err, invite.ErrTokenExpired)
	})

	t.Run("exceeding max uses returns error", func(t *testing.T) {
		mgr := invite.NewTokenManager(testKey)
		tokStr, err := mgr.CreateToken("127.0.0.1:9000", []byte("network-key-12345"), 1*time.Hour, 1)
		require.NoError(t, err)

		// First use: ok
		_, err = mgr.ValidateAndUse(tokStr)
		require.NoError(t, err)

		// Second use: exceeds max uses
		_, err = mgr.ValidateAndUse(tokStr)
		require.ErrorIs(t, err, invite.ErrTokenMaxUsesExceeded)
	})

	t.Run("tampered token returns signature error", func(t *testing.T) {
		mgr := invite.NewTokenManager(testKey)
		tokStr, err := mgr.CreateToken("127.0.0.1:9000", []byte("network-key-12345"), 1*time.Hour, 1)
		require.NoError(t, err)

		// Modify character in encoded string
		tampered := tokStr[:len(tokStr)-4] + "AAAA"
		_, err = mgr.ValidateAndUse(tampered)
		require.Error(t, err)
	})
}
