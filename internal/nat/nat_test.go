package nat_test

import (
	"net"
	"testing"
	"time"

	"github.com/GioTld/aldea/internal/nat"
	"github.com/pion/stun/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startMockSTUNServer(t *testing.T) (net.PacketConn, string) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	go func() {
		buf := make([]byte, 1024)
		for {
			n, addr, err := conn.ReadFrom(buf)
			if err != nil {
				return
			}

			msg := &stun.Message{Raw: buf[:n]}
			if err := msg.Decode(); err != nil {
				continue
			}

			if msg.Type == stun.BindingRequest {
				udpAddr, ok := addr.(*net.UDPAddr)
				if !ok {
					continue
				}

				res, err := stun.Build(
					stun.BindingSuccess,
					&stun.XORMappedAddress{
						IP:   udpAddr.IP,
						Port: udpAddr.Port,
					},
				)
				if err != nil {
					continue
				}
				res.TransactionID = msg.TransactionID

				_, _ = conn.WriteTo(res.Raw, addr)
			}
		}
	}()

	return conn, conn.LocalAddr().String()
}

func TestSTUNResolver(t *testing.T) {
	t.Run("discover mapped address from mock stun server", func(t *testing.T) {
		serverConn, serverAddr := startMockSTUNServer(t)
		defer serverConn.Close()

		clientConn, err := net.ListenPacket("udp", "127.0.0.1:0")
		require.NoError(t, err)
		defer clientConn.Close()

		client := nat.NewClient(serverAddr)
		mapped, err := client.DiscoverMappedAddress(clientConn)
		require.NoError(t, err)
		assert.NotNil(t, mapped)
		assert.True(t, mapped.IP.IsLoopback())
		assert.Greater(t, mapped.Port, 0)
		assert.Contains(t, mapped.String(), "127.0.0.1:")
	})

	t.Run("timeout on unreachable stun server", func(t *testing.T) {
		clientConn, err := net.ListenPacket("udp", "127.0.0.1:0")
		require.NoError(t, err)
		defer clientConn.Close()

		client := nat.NewClientWithTimeout("127.0.0.1:1", 100*time.Millisecond)
		_, err = client.DiscoverMappedAddress(clientConn)
		assert.Error(t, err)
	})

	t.Run("default client constructor", func(t *testing.T) {
		client := nat.NewClient("")
		assert.NotNil(t, client)
	})
}
