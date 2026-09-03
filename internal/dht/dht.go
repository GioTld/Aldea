package dht

import (
	"bytes"
	"crypto/sha256"
	"sort"
	"sync"
	"time"
)

const (
	IDBytes    = 20
	BucketSize = 20
)

type NodeID [IDBytes]byte

func NewNodeID(input string) NodeID {
	hash := sha256.Sum256([]byte(input))
	var id NodeID
	copy(id[:], hash[:IDBytes])
	return id
}

func XOR(a, b NodeID) NodeID {
	var res NodeID
	for i := 0; i < IDBytes; i++ {
		res[i] = a[i] ^ b[i]
	}
	return res
}

func (id NodeID) PrefixLen(target NodeID) int {
	dist := XOR(id, target)
	for i := 0; i < IDBytes; i++ {
		b := dist[i]
		if b != 0 {
			for bit := 7; bit >= 0; bit-- {
				if (b & (1 << bit)) != 0 {
					return i*8 + (7 - bit)
				}
			}
		}
	}
	return IDBytes * 8
}

type Contact struct {
	ID       NodeID
	Address  string
	LastSeen time.Time
}

type bucket struct {
	contacts []Contact
}

type RoutingTable struct {
	mu      sync.RWMutex
	localID NodeID
	buckets [IDBytes * 8]*bucket
}

func NewRoutingTable(localID NodeID) *RoutingTable {
	rt := &RoutingTable{localID: localID}
	for i := range rt.buckets {
		rt.buckets[i] = &bucket{}
	}
	return rt
}

func (rt *RoutingTable) Add(c Contact) bool {
	if c.ID == rt.localID {
		return false
	}

	rt.mu.Lock()
	defer rt.mu.Unlock()

	bucketIdx := rt.localID.PrefixLen(c.ID)
	if bucketIdx >= len(rt.buckets) {
		bucketIdx = len(rt.buckets) - 1
	}

	b := rt.buckets[bucketIdx]
	for i, existing := range b.contacts {
		if existing.ID == c.ID {
			b.contacts[i] = c
			return true
		}
	}

	if len(b.contacts) < BucketSize {
		b.contacts = append(b.contacts, c)
		return true
	}

	return false
}

func (rt *RoutingTable) Remove(id NodeID) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	bucketIdx := rt.localID.PrefixLen(id)
	if bucketIdx >= len(rt.buckets) {
		bucketIdx = len(rt.buckets) - 1
	}

	b := rt.buckets[bucketIdx]
	for i, existing := range b.contacts {
		if existing.ID == id {
			b.contacts = append(b.contacts[:i], b.contacts[i+1:]...)
			return
		}
	}
}

func (rt *RoutingTable) FindClosest(target NodeID, count int) []Contact {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	var all []Contact
	for _, b := range rt.buckets {
		all = append(all, b.contacts...)
	}

	sort.Slice(all, func(i, j int) bool {
		distI := XOR(all[i].ID, target)
		distJ := XOR(all[j].ID, target)
		return bytes.Compare(distI[:], distJ[:]) < 0
	})

	if len(all) > count {
		return all[:count]
	}
	return all
}
