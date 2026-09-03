package bandwidth

import (
	"io"
	"sync"
	"time"
)

// Limiter manages token bucket rate limiting for byte stream operations.
type Limiter struct {
	bytesPerSec int64
	mu          sync.Mutex
	lastCheck   time.Time
	tokens      float64
}

func NewLimiter(bytesPerSec int64) *Limiter {
	return &Limiter{
		bytesPerSec: bytesPerSec,
		lastCheck:   time.Now(),
		tokens:      float64(bytesPerSec),
	}
}

// WaitN blocks until n bytes are allowed to be processed according to the rate limit.
func (l *Limiter) WaitN(n int) {
	if l.bytesPerSec <= 0 || n <= 0 {
		return
	}

	rem := n
	// Process in chunk sizes no larger than bytesPerSec to handle large read/write buffers
	for rem > 0 {
		chunk := rem
		if int64(chunk) > l.bytesPerSec {
			chunk = int(l.bytesPerSec)
		}
		l.waitChunk(chunk)
		rem -= chunk
	}
}

func (l *Limiter) waitChunk(n int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	maxCapacity := float64(l.bytesPerSec)
	for {
		now := time.Now()
		elapsed := now.Sub(l.lastCheck).Seconds()
		l.lastCheck = now

		l.tokens += elapsed * float64(l.bytesPerSec)
		if l.tokens > maxCapacity {
			l.tokens = maxCapacity
		}

		if l.tokens >= float64(n) {
			l.tokens -= float64(n)
			return
		}

		needed := float64(n) - l.tokens
		sleepSec := needed / float64(l.bytesPerSec)
		l.mu.Unlock()
		time.Sleep(time.Duration(sleepSec * float64(time.Second)))
		l.mu.Lock()
	}
}

// ThrottledReader wraps an io.Reader and restricts its read throughput.
type ThrottledReader struct {
	r       io.Reader
	limiter *Limiter
}

func NewThrottledReader(r io.Reader, bytesPerSec int64) *ThrottledReader {
	return &ThrottledReader{
		r:       r,
		limiter: NewLimiter(bytesPerSec),
	}
}

func (tr *ThrottledReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	tr.limiter.WaitN(len(p))
	return tr.r.Read(p)
}

// ThrottledWriter wraps an io.Writer and restricts its write throughput.
type ThrottledWriter struct {
	w       io.Writer
	limiter *Limiter
}

func NewThrottledWriter(w io.Writer, bytesPerSec int64) *ThrottledWriter {
	return &ThrottledWriter{
		w:       w,
		limiter: NewLimiter(bytesPerSec),
	}
}

func (tw *ThrottledWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	tw.limiter.WaitN(len(p))
	return tw.w.Write(p)
}
