package dht_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/GioTld/aldea/internal/dht"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeIDAndXOR(t *testing.T) {
	id1 := dht.NewNodeID("node-1")
	id2 := dht.NewNodeID("node-2")
	id1Copy := dht.NewNodeID("node-1")

	assert.Equal(t, id1, id1Copy)
	assert.NotEqual(t, id1, id2)

	distSelf := dht.XOR(id1, id1)
	assert.Equal(t, dht.NodeID{}, distSelf)

	dist12 := dht.XOR(id1, id2)
	dist21 := dht.XOR(id2, id1)
	assert.Equal(t, dist12, dist21)
}

func TestRoutingTable(t *testing.T) {
	localID := dht.NewNodeID("local-node")
	rt := dht.NewRoutingTable(localID)

	t.Run("add and find closest contacts", func(t *testing.T) {
		for i := 1; i <= 25; i++ {
			nodeID := dht.NewNodeID(fmt.Sprintf("peer-node-%d", i))
			contact := dht.Contact{
				ID:       nodeID,
				Address:  fmt.Sprintf("192.168.1.%d:9000", i),
				LastSeen: time.Now(),
			}
			rt.Add(contact)
		}

		targetID := dht.NewNodeID("peer-node-5")
		closest := rt.FindClosest(targetID, 5)
		require.NotEmpty(t, closest)
		assert.LessOrEqual(t, len(closest), 5)
		assert.Equal(t, targetID, closest[0].ID)
	})

	t.Run("remove contact", func(t *testing.T) {
		peerID := dht.NewNodeID("remove-target")
		contact := dht.Contact{
			ID:       peerID,
			Address:  "10.0.0.99:9000",
			LastSeen: time.Now(),
		}
		rt.Add(contact)

		closestBefore := rt.FindClosest(peerID, 1)
		require.NotEmpty(t, closestBefore)
		assert.Equal(t, peerID, closestBefore[0].ID)

		rt.Remove(peerID)
		closestAfter := rt.FindClosest(peerID, 1)
		if len(closestAfter) > 0 {
			assert.NotEqual(t, peerID, closestAfter[0].ID)
		}
	})
}
