// Command stuk is the client: it sends the port-knock sequence, then an auth
// packet carrying a TOTP code (and optionally an SSH public key to provision).
package main

import (
	"flag"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/trustsentinel/stuk/internal/knock"
	"github.com/trustsentinel/stuk/pkg/crypto"
)

func main() {
	host := flag.String("host", "127.0.0.1", "target host")
	portsCSV := flag.String("ports", "", "knock sequence, comma-separated (e.g. 4000,4001,4002)")
	authPort := flag.Int("auth-port", 0, "auth port (required)")
	secret := flag.String("secret", "", "TOTP secret; the client derives the code from it")
	code := flag.String("code", "", "TOTP code (use instead of -secret)")
	pubkey := flag.String("pubkey", "", "SSH public key to provision")
	issuer := flag.String("issuer", "stuk", "TOTP issuer")
	gapMs := flag.Int("gap-ms", 200, "gap between knocks in milliseconds")
	flag.Parse()

	ports, err := parsePorts(*portsCSV)
	if err != nil || len(ports) == 0 {
		log.Fatalf("invalid -ports %q: %v", *portsCSV, err)
	}
	if *authPort == 0 {
		log.Fatal("-auth-port is required")
	}

	c := *code
	if c == "" {
		if *secret == "" {
			log.Fatal("provide -code or -secret")
		}
		c, err = crypto.NewTOTPManager(*issuer).GenerateCode(*secret)
		if err != nil {
			log.Fatalf("generate TOTP code: %v", err)
		}
	}

	log.Printf("knocking %s: %v", *host, ports)
	if err := knock.SendSequence(*host, ports, time.Duration(*gapMs)*time.Millisecond); err != nil {
		log.Fatalf("send knocks: %v", err)
	}
	if err := knock.SendAuth(*host, *authPort, knock.AuthRequest{Code: c, PubKey: *pubkey}); err != nil {
		log.Fatalf("send auth: %v", err)
	}
	log.Printf("knock sequence + auth sent to %s:%d", *host, *authPort)
}

func parsePorts(csv string) ([]int, error) {
	var out []int
	for _, s := range strings.Split(csv, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		p, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}
