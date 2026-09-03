package chunker

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
)

const DefaultChunkSize = 1024 * 1024

var (
	ErrInvalidChunkSequence = errors.New("chunks are out of order or missing sequence index")
	ErrCorruptedChunk        = errors.New("chunk checksum mismatch")
)

type Chunk struct {
	Index int
	Data  []byte
	Hash  [32]byte
}

func Split(r io.Reader, chunkSize int) ([]Chunk, error) {
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}

	var chunks []Chunk
	index := 0
	buf := make([]byte, chunkSize)

	for {
		n, err := io.ReadFull(r, buf)
		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])
			hash := sha256.Sum256(data)

			chunks = append(chunks, Chunk{
				Index: index,
				Data:  data,
				Hash:  hash,
			})
			index++
		}

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading stream: %w", err)
		}
	}

	return chunks, nil
}

func Reassemble(chunks []Chunk, w io.Writer) error {
	for i, c := range chunks {
		if c.Index != i {
			return ErrInvalidChunkSequence
		}

		if sha256.Sum256(c.Data) != c.Hash {
			return ErrCorruptedChunk
		}

		if _, err := w.Write(c.Data); err != nil {
			return fmt.Errorf("writing chunk %d: %w", i, err)
		}
	}

	return nil
}
