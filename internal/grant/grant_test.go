package grant

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestIptablesGranterArgs(t *testing.T) {
	var calls [][]string
	g := IptablesGranter{
		Run: func(name string, args ...string) error {
			calls = append(calls, append([]string{name}, args...))
			return nil
		},
	}
	if err := g.Grant("203.0.113.7", "ignored"); err != nil {
		t.Fatal(err)
	}
	if err := g.Revoke("203.0.113.7"); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"iptables -I INPUT -p tcp --dport 22 -s 203.0.113.7 -j ACCEPT",
		"iptables -D INPUT -p tcp --dport 22 -s 203.0.113.7 -j ACCEPT",
	}
	if len(calls) != 2 {
		t.Fatalf("got %d calls, want 2", len(calls))
	}
	for i, w := range want {
		if got := strings.Join(calls[i], " "); got != w {
			t.Errorf("call %d = %q, want %q", i, got, w)
		}
	}
}

func TestIptablesGranterCustomPortChain(t *testing.T) {
	var got string
	g := IptablesGranter{SSHPort: 2222, Chain: "STUK", Run: func(name string, args ...string) error {
		got = name + " " + strings.Join(args, " ")
		return nil
	}}
	if err := g.Grant("10.0.0.1", ""); err != nil {
		t.Fatal(err)
	}
	want := "iptables -I STUK -p tcp --dport 2222 -s 10.0.0.1 -j ACCEPT"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAuthKeysGranterStore(t *testing.T) {
	dir := t.TempDir()
	g := AuthKeysGranter{Dir: dir}
	const key = "ssh-ed25519 AAAAC3Nz... demo"

	if err := g.Grant("192.0.2.5", key); err != nil {
		t.Fatal(err)
	}
	keys, err := AuthorizedKeys(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || keys[0] != key {
		t.Fatalf("authorized keys = %v, want [%q]", keys, key)
	}

	if err := g.Revoke("192.0.2.5"); err != nil {
		t.Fatal(err)
	}
	keys, _ = AuthorizedKeys(dir)
	if len(keys) != 0 {
		t.Fatalf("after revoke, keys = %v, want none", keys)
	}
	// Revoke of an unknown IP is a no-op, not an error.
	if err := g.Revoke("192.0.2.9"); err != nil {
		t.Errorf("revoke of absent grant errored: %v", err)
	}
}

func TestAuthKeysGranterRejectsEmptyKey(t *testing.T) {
	g := AuthKeysGranter{Dir: t.TempDir()}
	if err := g.Grant("192.0.2.1", "   "); err == nil {
		t.Fatal("expected an error granting with an empty pubkey")
	}
}

func TestAuthorizedKeysMissingDir(t *testing.T) {
	keys, err := AuthorizedKeys(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil || keys != nil {
		t.Fatalf("missing dir: keys=%v err=%v, want nil,nil", keys, err)
	}
}

// recordGranter records Grant/Revoke calls for fan-out and TTL tests.
type recordGranter struct {
	mu      sync.Mutex
	grants  []string
	revokes []string
}

func (r *recordGranter) Grant(ip, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.grants = append(r.grants, ip)
	return nil
}
func (r *recordGranter) Revoke(ip string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.revokes = append(r.revokes, ip)
	return nil
}
func (r *recordGranter) revokeCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.revokes)
}

func TestMultiGranterFanout(t *testing.T) {
	a, b := &recordGranter{}, &recordGranter{}
	m := MultiGranter{a, b}
	if err := m.Grant("198.51.100.2", "k"); err != nil {
		t.Fatal(err)
	}
	if err := m.Revoke("198.51.100.2"); err != nil {
		t.Fatal(err)
	}
	if len(a.grants) != 1 || len(b.grants) != 1 || len(a.revokes) != 1 || len(b.revokes) != 1 {
		t.Fatalf("fan-out incomplete: a=%+v b=%+v", a, b)
	}
}

func TestManagerAutoRevokesAfterTTL(t *testing.T) {
	r := &recordGranter{}
	m := NewManager(r, 30*time.Millisecond)
	if err := m.Grant("192.0.2.50", "k"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if r.revokeCount() == 1 {
			return // auto-revoked
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("grant was not auto-revoked after the TTL")
}

func TestLogGranterIsNoError(t *testing.T) {
	// Ensure the default backend never errors (used as the safe fallback).
	var g Granter = LogGranter{}
	if err := g.Grant("192.0.2.1", ""); err != nil {
		t.Fatal(err)
	}
	if err := g.Revoke("192.0.2.1"); err != nil {
		t.Fatal(err)
	}
}
