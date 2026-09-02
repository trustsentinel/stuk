// Command stukd is the stuk port-knocking daemon: it watches for an ordered
// UDP knock sequence, then grants temporary access to sources that present a
// valid TOTP code — auto-revoking after a TTL.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/trustsentinel/stuk/internal/config"
	"github.com/trustsentinel/stuk/internal/grant"
	"github.com/trustsentinel/stuk/internal/knock"
	"github.com/trustsentinel/stuk/pkg/crypto"
)

func main() {
	cfgPath := flag.String("config", "stukd.json", "path to JSON config")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	tracker := knock.NewTracker(cfg.KnockPorts, cfg.Window())

	var g grant.Granter = grant.LogGranter{}
	if cfg.GrantMode == "script" {
		g = grant.ScriptGranter{GrantCmd: cfg.GrantCmd, RevokeCmd: cfg.RevokeCmd}
	}
	mgr := grant.NewManager(g, cfg.TTL())
	totpMgr := crypto.NewTOTPManager(cfg.Issuer)
	armed := &armedSet{m: make(map[string]time.Time), window: cfg.Window()}

	var wg sync.WaitGroup
	for _, p := range cfg.KnockPorts {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			listenUDP(cfg.BindAddr, port, func(ip string, _ []byte) {
				if tracker.Knock(ip, port, time.Now()) {
					armed.arm(ip)
					log.Printf("sequence complete from %s (awaiting auth)", ip)
				}
			})
		}(p)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		listenUDP(cfg.BindAddr, cfg.AuthPort, func(ip string, data []byte) {
			var req knock.AuthRequest
			if err := json.Unmarshal(data, &req); err != nil {
				return
			}
			if !armed.valid(ip) {
				log.Printf("auth from %s rejected: no completed sequence", ip)
				return
			}
			if !totpMgr.ValidateCodeWithWindow(cfg.TOTPSecret, req.Code, cfg.TOTPWindow) {
				log.Printf("auth from %s rejected: invalid TOTP", ip)
				return
			}
			armed.clear(ip)
			if err := mgr.Grant(ip, req.PubKey); err != nil {
				log.Printf("grant %s failed: %v", ip, err)
				return
			}
			log.Printf("ACCESS GRANTED to %s for %s", ip, cfg.TTL())
		})
	}()

	log.Printf("stukd listening on %s: knock=%v auth=%d window=%s ttl=%s grant=%s",
		cfg.BindAddr, cfg.KnockPorts, cfg.AuthPort, cfg.Window(), cfg.TTL(), cfg.GrantMode)
	wg.Wait()
}

// armedSet tracks IPs that just completed the knock sequence and may now auth.
type armedSet struct {
	mu     sync.Mutex
	m      map[string]time.Time
	window time.Duration
}

func (a *armedSet) arm(ip string) {
	a.mu.Lock()
	a.m[ip] = time.Now()
	a.mu.Unlock()
}

func (a *armedSet) valid(ip string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	t, ok := a.m[ip]
	return ok && time.Since(t) <= a.window
}

func (a *armedSet) clear(ip string) {
	a.mu.Lock()
	delete(a.m, ip)
	a.mu.Unlock()
}

func listenUDP(bind string, port int, handler func(ip string, data []byte)) {
	addr := net.JoinHostPort(bind, strconv.Itoa(port))
	pc, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Fatalf("listen %s: %v", addr, err)
	}
	defer pc.Close()
	buf := make([]byte, 4096)
	for {
		n, remote, err := pc.ReadFrom(buf)
		if err != nil {
			continue
		}
		host, _, splitErr := net.SplitHostPort(remote.String())
		if splitErr != nil {
			host = remote.String()
		}
		data := make([]byte, n)
		copy(data, buf[:n])
		handler(host, data)
	}
}
