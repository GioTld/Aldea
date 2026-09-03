package ingress_test

import (
	"context"
	"testing"

	"github.com/GioTld/aldea/internal/ingress"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIngressManager(t *testing.T) {
	ctx := context.Background()
	mgr := ingress.NewManager()

	t.Run("creates ingress tunnel for compute workload behind NAT (RF-31)", func(t *testing.T) {
		tunnel, err := mgr.CreateTunnel(ctx, "wl-web-8080", 8080, "relay.aldea.net:9090")
		require.NoError(t, err)
		assert.Equal(t, "wl-web-8080", tunnel.WorkloadID)
		assert.Equal(t, 8080, tunnel.TargetPort)
		assert.True(t, tunnel.IsActive)
		assert.NotEmpty(t, tunnel.PublicAddr)
	})

	t.Run("lists and closes active ingress tunnels", func(t *testing.T) {
		tunnel, err := mgr.CreateTunnel(ctx, "wl-api-3000", 3000, "relay.aldea.net:9090")
		require.NoError(t, err)

		tunnels, err := mgr.ListTunnels(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, tunnels)

		err = mgr.CloseTunnel(ctx, tunnel.TunnelID)
		require.NoError(t, err)
	})
}
