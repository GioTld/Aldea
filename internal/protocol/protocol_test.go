package protocol_test

import (
	"encoding/json"
	"testing"

	"github.com/GioTld/aldea/internal/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecode(t *testing.T) {
	netKey := []byte("secret-network-key-32-bytes-long!")
	nodeID := "node-alpha-1"

	t.Run("PutShardRequest round trip", func(t *testing.T) {
		req := protocol.PutShardRequest{
			ShardID: "shard-101",
			Data:    []byte("encrypted-shard-payload"),
		}

		encoded, err := protocol.Encode(protocol.TypePutShardRequest, nodeID, req, netKey)
		require.NoError(t, err)
		assert.NotEmpty(t, encoded)

		env, err := protocol.Decode(encoded, netKey)
		require.NoError(t, err)
		assert.Equal(t, protocol.TypePutShardRequest, env.Type)
		assert.Equal(t, nodeID, env.NodeID)

		var decodedReq protocol.PutShardRequest
		err = env.UnmarshalBody(&decodedReq)
		require.NoError(t, err)
		assert.Equal(t, req, decodedReq)
	})

	t.Run("PutShardResponse round trip", func(t *testing.T) {
		resp := protocol.PutShardResponse{
			ShardID: "shard-101",
			Success: true,
			Error:   "",
		}

		encoded, err := protocol.Encode(protocol.TypePutShardResponse, nodeID, resp, netKey)
		require.NoError(t, err)

		env, err := protocol.Decode(encoded, netKey)
		require.NoError(t, err)
		assert.Equal(t, protocol.TypePutShardResponse, env.Type)

		var decodedResp protocol.PutShardResponse
		err = env.UnmarshalBody(&decodedResp)
		require.NoError(t, err)
		assert.Equal(t, resp, decodedResp)
	})

	t.Run("GetShardRequest round trip", func(t *testing.T) {
		req := protocol.GetShardRequest{ShardID: "shard-202"}
		encoded, err := protocol.Encode(protocol.TypeGetShardRequest, nodeID, req, netKey)
		require.NoError(t, err)

		env, err := protocol.Decode(encoded, netKey)
		require.NoError(t, err)

		var decodedReq protocol.GetShardRequest
		err = env.UnmarshalBody(&decodedReq)
		require.NoError(t, err)
		assert.Equal(t, req, decodedReq)
	})

	t.Run("GetShardResponse round trip", func(t *testing.T) {
		resp := protocol.GetShardResponse{
			ShardID: "shard-202",
			Data:    []byte("downloaded-data"),
			Found:   true,
		}
		encoded, err := protocol.Encode(protocol.TypeGetShardResponse, nodeID, resp, netKey)
		require.NoError(t, err)

		env, err := protocol.Decode(encoded, netKey)
		require.NoError(t, err)

		var decodedResp protocol.GetShardResponse
		err = env.UnmarshalBody(&decodedResp)
		require.NoError(t, err)
		assert.Equal(t, resp, decodedResp)
	})

	t.Run("HealthReportRequest and Response round trip", func(t *testing.T) {
		req := protocol.HealthReportRequest{NodeID: nodeID}
		encodedReq, err := protocol.Encode(protocol.TypeHealthReportRequest, nodeID, req, netKey)
		require.NoError(t, err)

		envReq, err := protocol.Decode(encodedReq, netKey)
		require.NoError(t, err)
		assert.Equal(t, protocol.TypeHealthReportRequest, envReq.Type)

		resp := protocol.HealthReportResponse{
			NodeID:           nodeID,
			StorageUsed:      102400,
			StorageAllocated: 1048576,
			IsHealthy:        true,
		}
		encodedResp, err := protocol.Encode(protocol.TypeHealthReportResponse, nodeID, resp, netKey)
		require.NoError(t, err)

		envResp, err := protocol.Decode(encodedResp, netKey)
		require.NoError(t, err)

		var decodedResp protocol.HealthReportResponse
		err = envResp.UnmarshalBody(&decodedResp)
		require.NoError(t, err)
		assert.Equal(t, resp, decodedResp)
	})

	t.Run("tampered payload fails signature verification", func(t *testing.T) {
		req := protocol.GetShardRequest{ShardID: "shard-303"}
		encoded, err := protocol.Encode(protocol.TypeGetShardRequest, nodeID, req, netKey)
		require.NoError(t, err)

		var env protocol.MessageEnvelope
		require.NoError(t, json.Unmarshal(encoded, &env))
		env.Payload[0] ^= 0xff
		tampered, err := json.Marshal(env)
		require.NoError(t, err)

		_, err = protocol.Decode(tampered, netKey)
		assert.ErrorIs(t, err, protocol.ErrInvalidSignature)
	})

	t.Run("wrong network key fails authentication", func(t *testing.T) {
		req := protocol.GetShardRequest{ShardID: "shard-303"}
		encoded, err := protocol.Encode(protocol.TypeGetShardRequest, nodeID, req, netKey)
		require.NoError(t, err)

		wrongKey := []byte("unauthorized-network-key-32bytes")
		_, err = protocol.Decode(encoded, wrongKey)
		assert.ErrorIs(t, err, protocol.ErrInvalidSignature)
	})

	t.Run("malformed data fails decoding", func(t *testing.T) {
		_, err := protocol.Decode([]byte("bad json payload"), netKey)
		assert.Error(t, err)
	})
}
