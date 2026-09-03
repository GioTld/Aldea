package relay_test

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/GioTld/aldea/internal/relay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testKey = []byte("aldea-test-network-key-32bytes!!")

func TestFallbackDialer(t *testing.T) {
	t.Run("direct connection succeeds when target is reachable", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer ln.Close()

		go func() {
			c, _ := ln.Accept()
			if c != nil {
				io.Copy(c, c)
				c.Close()
			}
		}()

		dialer := relay.NewFallbackDialer(testKey)
		conn, usedRelay, err := dialer.Dial(ln.Addr().String(), "", 2*time.Second)
		require.NoError(t, err)
		defer conn.Close()

		assert.False(t, usedRelay)
		assert.NotNil(t, conn)
	})

	t.Run("falls back to relay when direct connection fails", func(t *testing.T) {
		srv := relay.NewServer(testKey)
		relayLn, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		go srv.Serve(relayLn)
		defer relayLn.Close()

		targetLn, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		go func() {
			c, _ := targetLn.Accept()
			if c != nil {
				io.Copy(c, c)
				c.Close()
			}
		}()
		defer targetLn.Close()

		dialer := relay.NewFallbackDialer(testKey)
		conn, usedRelay, err := dialer.Dial(
			"127.0.0.1:1",
			relayLn.Addr().String()+"?target="+targetLn.Addr().String(),
			500*time.Millisecond,
		)
		require.NoError(t, err)
		defer conn.Close()

		assert.True(t, usedRelay)

		payload := []byte("aldea relay test")
		_, err = conn.Write(payload)
		require.NoError(t, err)

		buf := make([]byte, len(payload))
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, err = io.ReadFull(conn, buf)
		require.NoError(t, err)
		assert.Equal(t, payload, buf)
	})
}

func TestRelayServer(t *testing.T) {
	t.Run("relay server proxies data between two connections", func(t *testing.T) {
		srv := relay.NewServer(testKey)
		relayLn, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		go srv.Serve(relayLn)
		defer relayLn.Close()

		targetLn, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer targetLn.Close()

		targetDone := make(chan struct{})
		go func() {
			defer close(targetDone)
			c, err := targetLn.Accept()
			if err != nil {
				return
			}
			defer c.Close()
			io.Copy(c, c)
		}()

		client := relay.NewClient(testKey)
		conn, err := client.CreateSession(
			relayLn.Addr().String(),
			targetLn.Addr().String(),
		)
		require.NoError(t, err)
		defer conn.Close()

		msg := []byte("hello through relay")
		_, err = conn.Write(msg)
		require.NoError(t, err)

		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		buf := make([]byte, len(msg))
		_, err = io.ReadFull(conn, buf)
		require.NoError(t, err)
		assert.Equal(t, msg, buf)
	})
}
