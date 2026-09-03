package invite

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrTokenExpired         = errors.New("invite token has expired")
	ErrTokenMaxUsesExceeded = errors.New("invite token usage limit reached")
	ErrInvalidSignature     = errors.New("invite token signature verification failed")
	ErrMalformedToken       = errors.New("malformed invite token")
)

type InviteToken struct {
	TokenID     string `json:"token_id"`
	TrackerAddr string `json:"tracker_addr"`
	NetworkKey  []byte `json:"network_key"`
	ExpiresAt   int64  `json:"expires_at"`
	MaxUses     int    `json:"max_uses"`
	MAC         []byte `json:"mac"`
}

type TokenUsage struct {
	UsedCount int
}

type TokenManager struct {
	signingKey []byte
	mu         sync.Mutex
	usage      map[string]*TokenUsage
}

func NewTokenManager(signingKey []byte) *TokenManager {
	return &TokenManager{
		signingKey: signingKey,
		usage:      make(map[string]*TokenUsage),
	}
}

func (m *TokenManager) CreateToken(trackerAddr string, networkKey []byte, ttl time.Duration, maxUses int) (string, error) {
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return "", fmt.Errorf("generating token id: %w", err)
	}
	tokenID := hex.EncodeToString(idBytes)

	expiresAt := time.Now().Add(ttl).Unix()
	if ttl <= 0 {
		expiresAt = time.Now().Add(ttl).Unix()
	}

	tok := InviteToken{
		TokenID:     tokenID,
		TrackerAddr: trackerAddr,
		NetworkKey:  networkKey,
		ExpiresAt:   expiresAt,
		MaxUses:     maxUses,
	}

	tok.MAC = m.computeMAC(tok)

	data, err := json.Marshal(tok)
	if err != nil {
		return "", fmt.Errorf("marshaling token: %w", err)
	}

	encoded := base64.URLEncoding.EncodeToString(data)
	return "aldea1_" + encoded, nil
}

func (m *TokenManager) ValidateAndUse(tokenStr string) (*InviteToken, error) {
	if len(tokenStr) < 7 || tokenStr[:7] != "aldea1_" {
		return nil, ErrMalformedToken
	}

	data, err := base64.URLEncoding.DecodeString(tokenStr[7:])
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMalformedToken, err)
	}

	var tok InviteToken
	if err := json.Unmarshal(data, &tok); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMalformedToken, err)
	}

	// Verify HMAC signature
	expectedMAC := m.computeMAC(tok)
	if !hmac.Equal(tok.MAC, expectedMAC) {
		return nil, ErrInvalidSignature
	}

	// Check expiration
	if time.Now().Unix() > tok.ExpiresAt {
		return nil, ErrTokenExpired
	}

	// Check usage count
	m.mu.Lock()
	defer m.mu.Unlock()

	u, exists := m.usage[tok.TokenID]
	if !exists {
		u = &TokenUsage{UsedCount: 0}
		m.usage[tok.TokenID] = u
	}

	if tok.MaxUses > 0 && u.UsedCount >= tok.MaxUses {
		return nil, ErrTokenMaxUsesExceeded
	}

	u.UsedCount++
	return &tok, nil
}

func (m *TokenManager) computeMAC(tok InviteToken) []byte {
	h := hmac.New(sha256.New, m.signingKey)
	fmt.Fprintf(h, "%s:%s:%d:%d:", tok.TokenID, tok.TrackerAddr, tok.ExpiresAt, tok.MaxUses)
	h.Write(tok.NetworkKey)
	return h.Sum(nil)
}
