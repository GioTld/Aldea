package erasure_test

import (
	"bytes"
	"testing"

	"github.com/GioTld/aldea/internal/erasure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoder(t *testing.T) {
	t.Run("encode and reconstruct with all shards", func(t *testing.T) {
		enc, err := erasure.NewEncoder(4, 4)
		require.NoError(t, err)

		data := []byte("hello, Aldea erasure coding test data payload")
		shards, err := enc.Encode(data)
		require.NoError(t, err)
		assert.Len(t, shards, 8)

		recovered, err := enc.Reconstruct(shards, len(data))
		require.NoError(t, err)
		assert.Equal(t, data, recovered)
	})

	t.Run("reconstruct with m missing shards", func(t *testing.T) {
		enc, err := erasure.NewEncoder(4, 4)
		require.NoError(t, err)

		data := bytes.Repeat([]byte("storage-pool-data-test-block"), 100)
		shards, err := enc.Encode(data)
		require.NoError(t, err)
		assert.Len(t, shards, 8)

		available := make([]erasure.Shard, 8)
		copy(available, shards)

		available[0].Data = nil
		available[2].Data = nil
		available[5].Data = nil
		available[7].Data = nil

		recovered, err := enc.Reconstruct(available, len(data))
		require.NoError(t, err)
		assert.Equal(t, data, recovered)
	})

	t.Run("reconstruct fails with fewer than k shards", func(t *testing.T) {
		enc, err := erasure.NewEncoder(4, 4)
		require.NoError(t, err)

		data := []byte("sample payload for failure test")
		shards, err := enc.Encode(data)
		require.NoError(t, err)

		available := make([]erasure.Shard, 8)
		copy(available, shards)

		available[0].Data = nil
		available[1].Data = nil
		available[2].Data = nil
		available[3].Data = nil
		available[4].Data = nil

		_, err = enc.Reconstruct(available, len(data))
		assert.ErrorIs(t, err, erasure.ErrTooFewShards)
	})

	t.Run("invalid parameters return error", func(t *testing.T) {
		_, err := erasure.NewEncoder(0, 4)
		assert.ErrorIs(t, err, erasure.ErrInvalidShardCount)

		_, err = erasure.NewEncoder(4, 0)
		assert.ErrorIs(t, err, erasure.ErrInvalidShardCount)
	})

	t.Run("reconstruct handles empty data", func(t *testing.T) {
		enc, err := erasure.NewEncoder(4, 4)
		require.NoError(t, err)

		shards, err := enc.Encode([]byte{})
		require.NoError(t, err)

		recovered, err := enc.Reconstruct(shards, 0)
		require.NoError(t, err)
		assert.Empty(t, recovered)
	})
}
