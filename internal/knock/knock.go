// Package knock implements ordered port-knock sequence detection (server side)
// and sending (client side), plus the small auth payload exchanged after a
// successful sequence.
package knock

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

// AuthRequest is the payload sent to the auth port after the knock sequence
// completes. The TOTP code gates access; the public key is provisioned.
type AuthRequest struct {
	Code   string `json:"code"`
	PubKey string `json:"pubkey,omitempty"`
	User   string `json:"user,omitempty"`
}

// Tracker detects an ordered port-knock sequence, per source IP. It is safe for
// concurrent use by multiple UDP listeners.
type Tracker struct {
	seq    []int
	window time.Duration
	mu     sync.Mutex
	state  map[string]*progress
}

type progress struct {
	idx  int
	last time.Time
}

// NewTracker returns a Tracker for the given ordered sequence. A partial
// sequence resets if two knocks are more than window apart.
func NewTracker(seq []int, window time.Duration) *Tracker {
	return &Tracker{seq: seq, window: window, state: make(map[string]*progress)}
}

// Knock records a hit on port from ip at time now, and reports true exactly when
// the full sequence has just completed for that ip.
func (t *Tracker) Knock(ip string, port int, now time.Time) bool {
	if len(t.seq) == 0 {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	p := t.state[ip]
	if p == nil {
		p = &progress{}
		t.state[ip] = p
	}
	// Stale partial progress expires.
	if p.idx > 0 && now.Sub(p.last) > t.window {
		p.idx = 0
	}
	if port == t.seq[p.idx] {
		p.idx++
		p.last = now
		if p.idx == len(t.seq) {
			delete(t.state, ip) // completed; forget progress
			return true
		}
		return false
	}
	// Wrong port: restart. A hit on the first port begins a fresh attempt.
	if port == t.seq[0] {
		p.idx = 1
		p.last = now
	} else {
		p.idx = 0
	}
	return false
}

// SendSequence sends UDP knocks to host on each port in order, waiting gap
// between packets.
func SendSequence(host string, ports []int, gap time.Duration) error {
	for _, port := range ports {
		if err := sendUDP(host, port, []byte{0x01}); err != nil {
			return fmt.Errorf("knock %d: %w", port, err)
		}
		time.Sleep(gap)
	}
	return nil
}

// SendAuth sends the auth payload (TOTP code + optional key) to the auth port.
func SendAuth(host string, authPort int, req AuthRequest) error {
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return sendUDP(host, authPort, b)
}

func sendUDP(host string, port int, payload []byte) error {
	conn, err := net.Dial("udp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write(payload)
	return err
}
