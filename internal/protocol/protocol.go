package protocol

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type MessageType uint8

const (
	TypePutShardRequest MessageType = iota + 1
	TypePutShardResponse
	TypeGetShardRequest
	TypeGetShardResponse
	TypeHealthReportRequest
	TypeHealthReportResponse
	TypeRelaySessionRequest
	TypeRelaySessionResponse
)

var (
	ErrInvalidSignature       = errors.New("protocol authentication failed: signature mismatch")
	ErrUnsupportedMessageType = errors.New("unsupported protocol message type")
	ErrMalformedMessage       = errors.New("malformed protocol message format")
)

type PutShardRequest struct {
	ShardID string `json:"shard_id"`
	Data    []byte `json:"data"`
}

type PutShardResponse struct {
	ShardID string `json:"shard_id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type GetShardRequest struct {
	ShardID string `json:"shard_id"`
}

type GetShardResponse struct {
	ShardID string `json:"shard_id"`
	Data    []byte `json:"data,omitempty"`
	Found   bool   `json:"found"`
}

type HealthReportRequest struct {
	NodeID string `json:"node_id"`
}

type HealthReportResponse struct {
	NodeID           string `json:"node_id"`
	StorageUsed      int64  `json:"storage_used"`
	StorageAllocated int64  `json:"storage_allocated"`
	IsHealthy        bool   `json:"is_healthy"`
}

type RelaySessionRequest struct {
	SessionID    string `json:"session_id"`
	TargetNodeID string `json:"target_node_id"`
}

type RelaySessionResponse struct {
	SessionID string `json:"session_id"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

type MessageEnvelope struct {
	Type      MessageType `json:"type"`
	NodeID    string      `json:"node_id"`
	Timestamp int64       `json:"timestamp"`
	MAC       [32]byte    `json:"mac"`
	Payload   []byte      `json:"payload"`
}

func (e *MessageEnvelope) UnmarshalBody(target any) error {
	if err := json.Unmarshal(e.Payload, target); err != nil {
		return fmt.Errorf("%w: %v", ErrMalformedMessage, err)
	}
	return nil
}

func calculateMAC(msgType MessageType, nodeID string, timestamp int64, payload, networkKey []byte) [32]byte {
	h := hmac.New(sha256.New, networkKey)
	fmt.Fprintf(h, "%d:%s:%d:", msgType, nodeID, timestamp)
	h.Write(payload)

	var mac [32]byte
	copy(mac[:], h.Sum(nil))
	return mac
}

func Encode(msgType MessageType, nodeID string, body any, networkKey []byte) ([]byte, error) {
	if msgType < TypePutShardRequest || msgType > TypeHealthReportResponse {
		return nil, ErrUnsupportedMessageType
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling message body: %w", err)
	}

	timestamp := time.Now().UnixNano()
	mac := calculateMAC(msgType, nodeID, timestamp, payload, networkKey)

	env := MessageEnvelope{
		Type:      msgType,
		NodeID:    nodeID,
		Timestamp: timestamp,
		MAC:       mac,
		Payload:   payload,
	}

	return json.Marshal(env)
}

func Decode(data []byte, networkKey []byte) (*MessageEnvelope, error) {
	var env MessageEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMalformedMessage, err)
	}

	expectedMAC := calculateMAC(env.Type, env.NodeID, env.Timestamp, env.Payload, networkKey)
	if !hmac.Equal(env.MAC[:], expectedMAC[:]) {
		return nil, ErrInvalidSignature
	}

	return &env, nil
}
