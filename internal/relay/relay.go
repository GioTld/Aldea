package relay

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Server accepts relay requests and proxies data between the initiating
// caller and the target address they specify. No application-level framing is
// needed: once the target connection is established we simply pipe bytes in
// both directions.
type Server struct {
	networkKey []byte
}

func NewServer(networkKey []byte) *Server {
	return &Server{networkKey: networkKey}
}

func (s *Server) Serve(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go s.handleConn(conn)
	}
}

// handleConn reads a single line containing the target address, dials it, then
// pipes both connections bidirectionally.
func (s *Server) handleConn(src net.Conn) {
	defer src.Close()

	buf := make([]byte, 256)
	src.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := src.Read(buf)
	if err != nil {
		return
	}
	src.SetReadDeadline(time.Time{})

	targetAddr := strings.TrimSpace(string(buf[:n]))

	dst, err := net.DialTimeout("tcp", targetAddr, 5*time.Second)
	if err != nil {
		src.Write([]byte("ERR\n"))
		return
	}
	defer dst.Close()

	src.Write([]byte("OK\n"))
	pipe(src, dst)
}

func pipe(a, b net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); io.Copy(a, b) }()
	go func() { defer wg.Done(); io.Copy(b, a) }()
	wg.Wait()
}

// Client connects to a relay server and requests proxying to a target address.
type Client struct {
	networkKey []byte
}

func NewClient(networkKey []byte) *Client {
	return &Client{networkKey: networkKey}
}

// CreateSession connects to relayAddr and requests a proxied session to
// targetAddr. Returns a net.Conn representing the end-to-end channel.
func (c *Client) CreateSession(relayAddr, targetAddr string) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", relayAddr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connecting to relay server: %w", err)
	}

	_, err = fmt.Fprintf(conn, "%s\n", targetAddr)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("sending target address to relay: %w", err)
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	resp := make([]byte, 4)
	n, err := conn.Read(resp)
	conn.SetReadDeadline(time.Time{})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("reading relay handshake: %w", err)
	}

	if strings.TrimSpace(string(resp[:n])) != "OK" {
		conn.Close()
		return nil, fmt.Errorf("relay server rejected session to %s", targetAddr)
	}

	return conn, nil
}

// FallbackDialer tries a direct TCP connection first; if that fails or times
// out it transparently retries through the relay server.
type FallbackDialer struct {
	networkKey []byte
}

func NewFallbackDialer(networkKey []byte) *FallbackDialer {
	return &FallbackDialer{networkKey: networkKey}
}

// Dial attempts a direct TCP connection to directAddr. If that fails and
// relaySpec is non-empty it falls back to the relay. relaySpec uses the form
// "<relayHost:port>?target=<targetHost:port>"; the query parameter lets the
// caller override the target when the direct address is unreachable.
//
// Returns (conn, usedRelay, error).
func (d *FallbackDialer) Dial(directAddr, relaySpec string, timeout time.Duration) (net.Conn, bool, error) {
	conn, err := net.DialTimeout("tcp", directAddr, timeout)
	if err == nil {
		return conn, false, nil
	}

	if relaySpec == "" {
		return nil, false, fmt.Errorf("direct connection failed and no relay configured: %w", err)
	}

	relayAddr, targetAddr, parseErr := parseRelaySpec(relaySpec)
	if parseErr != nil {
		return nil, false, fmt.Errorf("invalid relay spec %q: %w", relaySpec, parseErr)
	}

	client := NewClient(d.networkKey)
	relayConn, relayErr := client.CreateSession(relayAddr, targetAddr)
	if relayErr != nil {
		return nil, false, fmt.Errorf("relay fallback failed: %w", relayErr)
	}

	return relayConn, true, nil
}

// parseRelaySpec splits "<relayAddr>?target=<targetAddr>".
func parseRelaySpec(spec string) (relayAddr, targetAddr string, err error) {
	parts := strings.SplitN(spec, "?", 2)
	relayAddr = parts[0]
	if len(parts) == 1 {
		return relayAddr, "", nil
	}

	params, err := url.ParseQuery(parts[1])
	if err != nil {
		return "", "", err
	}
	targetAddr = params.Get("target")
	return relayAddr, targetAddr, nil
}
