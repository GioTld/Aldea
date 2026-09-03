package erasure

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/klauspost/reedsolomon"
)

const (
	DefaultDataShards   = 4
	DefaultParityShards = 4
)

var (
	ErrInvalidShardCount   = errors.New("data and parity shard counts must be greater than zero")
	ErrTooFewShards        = errors.New("too few valid shards available to reconstruct original data")
	ErrReconstructionFailed = errors.New("erasure coding reconstruction failed")
)

type Shard struct {
	Index int
	Data  []byte
}

type Encoder struct {
	dataShards   int
	parityShards int
	rs           reedsolomon.Encoder
}

func NewEncoder(dataShards, parityShards int) (*Encoder, error) {
	if dataShards <= 0 || parityShards <= 0 {
		return nil, ErrInvalidShardCount
	}

	rs, err := reedsolomon.New(dataShards, parityShards)
	if err != nil {
		return nil, fmt.Errorf("creating reedsolomon encoder: %w", err)
	}

	return &Encoder{
		dataShards:   dataShards,
		parityShards: parityShards,
		rs:           rs,
	}, nil
}

func (e *Encoder) Encode(data []byte) ([]Shard, error) {
	if len(data) == 0 {
		shards := make([]Shard, e.dataShards+e.parityShards)
		for i := range shards {
			shards[i] = Shard{Index: i, Data: []byte{}}
		}
		return shards, nil
	}

	rawShards, err := e.rs.Split(data)
	if err != nil {
		return nil, fmt.Errorf("splitting data for erasure coding: %w", err)
	}

	if err := e.rs.Encode(rawShards); err != nil {
		return nil, fmt.Errorf("encoding parity shards: %w", err)
	}

	shards := make([]Shard, len(rawShards))
	for i, s := range rawShards {
		shards[i] = Shard{
			Index: i,
			Data:  s,
		}
	}

	return shards, nil
}

func (e *Encoder) Reconstruct(shards []Shard, originalSize int) ([]byte, error) {
	if originalSize == 0 {
		return []byte{}, nil
	}

	totalShards := e.dataShards + e.parityShards
	if len(shards) < totalShards {
		return nil, ErrTooFewShards
	}

	rawShards := make([][]byte, totalShards)
	validCount := 0
	for _, s := range shards {
		if s.Index >= 0 && s.Index < totalShards && len(s.Data) > 0 {
			rawShards[s.Index] = s.Data
			validCount++
		}
	}

	if validCount < e.dataShards {
		return nil, ErrTooFewShards
	}

	if err := e.rs.Reconstruct(rawShards); err != nil {
		return nil, errors.Join(ErrReconstructionFailed, err)
	}

	var buf bytes.Buffer
	if err := e.rs.Join(&buf, rawShards, originalSize); err != nil {
		return nil, fmt.Errorf("joining reconstructed shards: %w", err)
	}

	return buf.Bytes(), nil
}
