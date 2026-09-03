package chunker_test

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/GioTld/aldea/internal/chunker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitAndReassemble(t *testing.T) {
	t.Run("round trip with default chunk size", func(t *testing.T) {
		data := bytes.Repeat([]byte("abcdefghij"), 200000)
		r := bytes.NewReader(data)

		chunks, err := chunker.Split(r, 0)
		require.NoError(t, err)
		assert.NotEmpty(t, chunks)

		var buf bytes.Buffer
		err = chunker.Reassemble(chunks, &buf)
		require.NoError(t, err)
		assert.Equal(t, data, buf.Bytes())
	})

	t.Run("exact multiple of chunk size", func(t *testing.T) {
		chunkSize := 512
		data := bytes.Repeat([]byte("x"), chunkSize*4)
		r := bytes.NewReader(data)

		chunks, err := chunker.Split(r, chunkSize)
		require.NoError(t, err)
		assert.Len(t, chunks, 4)

		for i, c := range chunks {
			assert.Equal(t, i, c.Index)
			assert.Len(t, c.Data, chunkSize)
			assert.Equal(t, sha256.Sum256(c.Data), c.Hash)
		}

		var buf bytes.Buffer
		err = chunker.Reassemble(chunks, &buf)
		require.NoError(t, err)
		assert.Equal(t, data, buf.Bytes())
	})

	t.Run("non exact multiple has smaller last chunk", func(t *testing.T) {
		chunkSize := 512
		totalLen := chunkSize*2 + 100
		data := bytes.Repeat([]byte("y"), totalLen)
		r := bytes.NewReader(data)

		chunks, err := chunker.Split(r, chunkSize)
		require.NoError(t, err)
		require.Len(t, chunks, 3)

		assert.Len(t, chunks[0].Data, chunkSize)
		assert.Len(t, chunks[1].Data, chunkSize)
		assert.Len(t, chunks[2].Data, 100)

		var buf bytes.Buffer
		err = chunker.Reassemble(chunks, &buf)
		require.NoError(t, err)
		assert.Equal(t, data, buf.Bytes())
	})

	t.Run("empty reader produces no chunks", func(t *testing.T) {
		r := bytes.NewReader([]byte{})
		chunks, err := chunker.Split(r, 1024)
		require.NoError(t, err)
		assert.Empty(t, chunks)

		var buf bytes.Buffer
		err = chunker.Reassemble(chunks, &buf)
		require.NoError(t, err)
		assert.Empty(t, buf.Bytes())
	})

	t.Run("out of order chunks error on reassemble", func(t *testing.T) {
		data := bytes.Repeat([]byte("z"), 1500)
		chunks, err := chunker.Split(bytes.NewReader(data), 512)
		require.NoError(t, err)
		require.Len(t, chunks, 3)

		outOfOrder := []chunker.Chunk{chunks[1], chunks[0], chunks[2]}
		var buf bytes.Buffer
		err = chunker.Reassemble(outOfOrder, &buf)
		assert.ErrorIs(t, err, chunker.ErrInvalidChunkSequence)
	})

	t.Run("corrupted chunk hash fails reassemble", func(t *testing.T) {
		data := bytes.Repeat([]byte("a"), 1000)
		chunks, err := chunker.Split(bytes.NewReader(data), 512)
		require.NoError(t, err)

		chunks[0].Data[0] ^= 0xff
		var buf bytes.Buffer
		err = chunker.Reassemble(chunks, &buf)
		assert.ErrorIs(t, err, chunker.ErrCorruptedChunk)
	})
}

func TestChunkIndexFormat(t *testing.T) {
	chunk := chunker.Chunk{Index: 5}
	assert.Equal(t, "5", fmt.Sprint(chunk.Index))
}
