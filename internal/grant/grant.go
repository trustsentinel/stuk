// Package grant provisions temporary access after a successful knock+auth, and
// auto-revokes it after a TTL.
package grant

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Granter provisions and revokes access for a client IP.
type Granter interface {
	Grant(ip, pubkey string) error
	Revoke(ip string) error
}

// LogGranter is the safe default: it only logs. Useful for testing and dry-runs.
type LogGranter struct{}

func (LogGranter) Grant(ip, pubkey string) error { log.Printf("[grant] GRANT %s", ip); return nil }
func (LogGranter) Revoke(ip string) error        { log.Printf("[grant] REVOKE %s", ip); return nil }

// ScriptGranter runs shell commands with {ip} and {pubkey} placeholders — e.g.
// an iptables rule or an AuthorizedKeysCommand update. Empty commands are no-ops.
type ScriptGranter struct {
	GrantCmd  string
	RevokeCmd string
}

func (s ScriptGranter) run(tmpl, ip, pubkey string) error {
	if strings.TrimSpace(tmpl) == "" {
		return nil
	}
	cmd := strings.NewReplacer("{ip}", ip, "{pubkey}", pubkey).Replace(tmpl)
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (s ScriptGranter) Grant(ip, pubkey string) error { return s.run(s.GrantCmd, ip, pubkey) }
func (s ScriptGranter) Revoke(ip string) error        { return s.run(s.RevokeCmd, ip, "") }

// Manager wraps a Granter, granting for a fixed TTL and auto-revoking. A repeat
// grant for the same IP refreshes the TTL.
type Manager struct {
	g      Granter
	ttl    time.Duration
	mu     sync.Mutex
	timers map[string]*time.Timer
}

func NewManager(g Granter, ttl time.Duration) *Manager {
	return &Manager{g: g, ttl: ttl, timers: make(map[string]*time.Timer)}
}

func (m *Manager) Grant(ip, pubkey string) error {
	if err := m.g.Grant(ip, pubkey); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.timers[ip]; ok {
		t.Stop()
	}
	m.timers[ip] = time.AfterFunc(m.ttl, func() {
		if err := m.g.Revoke(ip); err != nil {
			log.Printf("[grant] revoke %s failed: %v", ip, err)
		}
		m.mu.Lock()
		delete(m.timers, ip)
		m.mu.Unlock()
	})
	return nil
}
