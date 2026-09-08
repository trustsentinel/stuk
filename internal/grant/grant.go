// Package grant provisions temporary access after a successful knock+auth, and
// auto-revokes it after a TTL.
package grant

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// IptablesGranter opens the SSH port for a client IP by inserting an ACCEPT rule
// (above the default DROP) on Grant, and deleting the same rule on Revoke. This
// is the native replacement for a grant/revoke shell script. Needs NET_ADMIN.
type IptablesGranter struct {
	SSHPort int    // default 22
	Chain   string // default "INPUT"
	Bin     string // default "iptables"
	// Run, if set, replaces exec (for tests). It receives the binary + args.
	Run func(name string, args ...string) error
}

func (g IptablesGranter) bin() string {
	if g.Bin != "" {
		return g.Bin
	}
	return "iptables"
}

func (g IptablesGranter) chain() string {
	if g.Chain != "" {
		return g.Chain
	}
	return "INPUT"
}

func (g IptablesGranter) port() int {
	if g.SSHPort != 0 {
		return g.SSHPort
	}
	return 22
}

// ruleArgs is the match for the client's SSH access (without the -I/-D verb).
func (g IptablesGranter) ruleArgs(ip string) []string {
	return []string{"-p", "tcp", "--dport", strconv.Itoa(g.port()), "-s", ip, "-j", "ACCEPT"}
}

func (g IptablesGranter) exec(args ...string) error {
	if g.Run != nil {
		return g.Run(g.bin(), args...)
	}
	out, err := exec.Command(g.bin(), args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w: %s", g.bin(), strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (g IptablesGranter) Grant(ip, _ string) error {
	return g.exec(append([]string{"-I", g.chain()}, g.ruleArgs(ip)...)...)
}

func (g IptablesGranter) Revoke(ip string) error {
	return g.exec(append([]string{"-D", g.chain()}, g.ruleArgs(ip)...)...)
}

// AuthKeysGranter time-gates the SSH key itself: on Grant it writes the client's
// public key into Dir (one file per IP); on Revoke it removes it. sshd is
// configured with AuthorizedKeysCommand=stuk-authkeys, which prints these keys —
// so a key is accepted only while a grant is active, in addition to the firewall.
type AuthKeysGranter struct {
	Dir string // default "/run/stuk/keys"
}

func (g AuthKeysGranter) dir() string {
	if g.Dir != "" {
		return g.Dir
	}
	return "/run/stuk/keys"
}

func (g AuthKeysGranter) file(ip string) string {
	safe := strings.NewReplacer("/", "_", ":", "_").Replace(ip)
	return filepath.Join(g.dir(), "ip-"+safe)
}

func (g AuthKeysGranter) Grant(ip, pubkey string) error {
	if strings.TrimSpace(pubkey) == "" {
		return fmt.Errorf("authkeys: no public key provided for %s", ip)
	}
	if err := os.MkdirAll(g.dir(), 0o755); err != nil {
		return err
	}
	return os.WriteFile(g.file(ip), []byte(strings.TrimSpace(pubkey)+"\n"), 0o644)
}

func (g AuthKeysGranter) Revoke(ip string) error {
	if err := os.Remove(g.file(ip)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// AuthorizedKeys returns the union of all currently-granted public keys in dir,
// for the stuk-authkeys command sshd calls. A missing dir yields no keys.
func AuthorizedKeys(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var keys []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "ip-") {
			continue
		}
		b, rerr := os.ReadFile(filepath.Join(dir, e.Name()))
		if rerr != nil {
			continue
		}
		for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
			if s := strings.TrimSpace(line); s != "" {
				keys = append(keys, s)
			}
		}
	}
	return keys, nil
}

// MultiGranter applies several granters in order — e.g. iptables + authkeys, to
// gate both the network and the key. Grant stops on the first error; Revoke is
// best-effort across all and returns the first error.
type MultiGranter []Granter

func (m MultiGranter) Grant(ip, pubkey string) error {
	for _, g := range m {
		if err := g.Grant(ip, pubkey); err != nil {
			return err
		}
	}
	return nil
}

func (m MultiGranter) Revoke(ip string) error {
	var firstErr error
	for _, g := range m {
		if err := g.Revoke(ip); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

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
