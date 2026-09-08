// Package config loads the stukd daemon configuration from a JSON file.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	BindAddr      string `json:"bind_addr"`
	KnockPorts    []int  `json:"knock_ports"`
	AuthPort      int    `json:"auth_port"`
	WindowSeconds int    `json:"window_seconds"`
	TTLSeconds    int    `json:"ttl_seconds"`
	Issuer        string `json:"issuer"`
	TOTPSecret    string `json:"totp_secret"`
	TOTPWindow    int    `json:"totp_window"`
	// GrantMode is a comma-separated list of backends applied together, any of:
	// "log", "script", "iptables", "authkeys".
	GrantMode     string `json:"grant_mode"`
	GrantCmd      string `json:"grant_cmd"`      // script mode
	RevokeCmd     string `json:"revoke_cmd"`     // script mode
	SSHPort       int    `json:"ssh_port"`       // iptables mode (default 22)
	IptablesChain string `json:"iptables_chain"` // iptables mode (default INPUT)
	AuthKeysDir   string `json:"authkeys_dir"`   // authkeys mode (default /run/stuk/keys)
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	c.applyDefaults()
	if err := c.validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) applyDefaults() {
	if c.BindAddr == "" {
		c.BindAddr = "0.0.0.0"
	}
	if c.WindowSeconds == 0 {
		c.WindowSeconds = 10
	}
	if c.TTLSeconds == 0 {
		c.TTLSeconds = 300
	}
	if c.Issuer == "" {
		c.Issuer = "stuk"
	}
	if c.TOTPWindow == 0 {
		c.TOTPWindow = 1
	}
	if c.GrantMode == "" {
		c.GrantMode = "log"
	}
}

func (c *Config) validate() error {
	if len(c.KnockPorts) == 0 {
		return fmt.Errorf("knock_ports must not be empty")
	}
	if c.AuthPort == 0 {
		return fmt.Errorf("auth_port is required")
	}
	if c.TOTPSecret == "" {
		return fmt.Errorf("totp_secret is required")
	}
	return nil
}

func (c *Config) Window() time.Duration { return time.Duration(c.WindowSeconds) * time.Second }
func (c *Config) TTL() time.Duration    { return time.Duration(c.TTLSeconds) * time.Second }

// Modes returns the configured grant backends (comma-separated), lowercased and
// trimmed; defaults to ["log"].
func (c *Config) Modes() []string {
	var modes []string
	for _, m := range strings.Split(c.GrantMode, ",") {
		if s := strings.ToLower(strings.TrimSpace(m)); s != "" {
			modes = append(modes, s)
		}
	}
	if len(modes) == 0 {
		return []string{"log"}
	}
	return modes
}
