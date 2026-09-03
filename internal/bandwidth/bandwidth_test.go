package bandwidth_test

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/GioTld/aldea/internal/bandwidth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThrottledReader(t *testing.T) {
	t.Run("unlimited bandwidth reads at full speed", func(t *testing.T) {
		data := []byte("hello bandwidth throttling test")
		r := bandwidth.NewThrottledReader(bytes.NewReader(data), 0)

		buf := make([]byte, len(data))
		n, err := r.Read(buf)
		require.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.Equal(t, data, buf)
	})

	t.Run("throttled reader paces reading according to limit", func(t *testing.T) {
		limitBytesPerSec := int64(2000)
		data := bytes.Repeat([]byte("A"), 3000)

		r := bandwidth.NewThrottledReader(bytes.NewReader(data), limitBytesPerSec)

		// Drain initial burst capacity
		drainBuf := make([]byte, 2000)
		r.Read(drainBuf)

		start := time.Now()
		buf := make([]byte, 1000)
		n, err := io.ReadFull(r, buf)
		elapsed := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, 1000, n)
		// 1000 bytes at 2000 B/s after empty bucket requires ~0.5s wait
		assert.GreaterOrEqual(t, elapsed, 300*time.Millisecond)
	})
}

func TestThrottledWriter(t *testing.T) {
	t.Run("unlimited bandwidth writes at full speed", func(t *testing.T) {
		var out bytes.Buffer
		w := bandwidth.NewThrottledWriter(&out, 0)

		data := []byte("hello bandwidth writer test")
		n, err := w.Write(data)
		require.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.Equal(t, data, out.Bytes())
	})

	t.Run("throttled writer paces writing according to limit", func(t *testing.T) {
		limitBytesPerSec := int64(2000)

		var out bytes.Buffer
		w := bandwidth.NewThrottledWriter(&out, limitBytesPerSec)

		// Drain initial burst capacity
		w.Write(bytes.Repeat([]byte("B"), 2000))

		start := time.Now()
		n, err := w.Write(bytes.Repeat([]byte("C"), 1000))
		elapsed := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, 1000, n)
		assert.GreaterOrEqual(t, elapsed, 300*time.Millisecond)
	})
}
