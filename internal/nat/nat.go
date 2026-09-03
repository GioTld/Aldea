package nat

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/pion/stun/v3"
)

const (
	DefaultSTUNServer = "stun.l.google.com:19302"
	DefaultTimeout    = 3 * time.Second
)

var (
	ErrSTUNTimeout          = errors.New("stun server request timed out")
	ErrInvalidSTUNResponse = errors.New("invalid or missing xor-mapped-address in stun response")
)

type MappedAddress struct {
	IP   net.IP
	Port int
}

func (m *MappedAddress) String() string {
	return net.JoinHostPort(m.IP.String(), fmt.Sprintf("%d", m.Port))
}

type Client struct {
	stunServer string
	timeout    time.Duration
}

func NewClient(stunServer string) *Client {
	if stunServer == "" {
		stunServer = DefaultSTUNServer
	}
	return &Client{
		stunServer: stunServer,
		timeout:    DefaultTimeout,
	}
}

func NewClientWithTimeout(stunServer string, timeout time.Duration) *Client {
	if stunServer == "" {
		stunServer = DefaultSTUNServer
	}
	return &Client{
		stunServer: stunServer,
		timeout:    timeout,
	}
}

func (c *Client) DiscoverMappedAddress(conn net.PacketConn) (*MappedAddress, error) {
	raddr, err := net.ResolveUDPAddr("udp", c.stunServer)
	if err != nil {
		return nil, fmt.Errorf("resolving stun server address: %w", err)
	}

	request, err := stun.Build(stun.TransactionID, stun.BindingRequest)
	if err != nil {
		return nil, fmt.Errorf("building stun request: %w", err)
	}

	if _, err := conn.WriteTo(request.Raw, raddr); err != nil {
		return nil, fmt.Errorf("sending stun request: %w", err)
	}

	udpConn, ok := conn.(*net.UDPConn)
	if ok {
		_ = udpConn.SetReadDeadline(time.Now().Add(c.timeout))
		defer func() { _ = udpConn.SetReadDeadline(time.Time{}) }()
	}

	buf := make([]byte, 1024)
	n, _, err := conn.ReadFrom(buf)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, ErrSTUNTimeout
		}
		return nil, fmt.Errorf("reading stun response: %w", err)
	}

	response := &stun.Message{Raw: buf[:n]}
	if err := response.Decode(); err != nil {
		return nil, fmt.Errorf("decoding stun response: %w", err)
	}

	var xorAddr stun.XORMappedAddress
	if err := xorAddr.GetFrom(response); err == nil {
		return &MappedAddress{IP: xorAddr.IP, Port: xorAddr.Port}, nil
	}

	var mappedAddr stun.MappedAddress
	if err := mappedAddr.GetFrom(response); err == nil {
		return &MappedAddress{IP: mappedAddr.IP, Port: mappedAddr.Port}, nil
	}

	return nil, ErrInvalidSTUNResponse
}

func (c *Client) GetPublicAddress() (*MappedAddress, error) {
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil, fmt.Errorf("listening local udp packet conn: %w", err)
	}
	defer conn.Close()

	return c.DiscoverMappedAddress(conn)
}
