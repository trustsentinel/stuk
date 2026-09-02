// Package config loads the stukd daemon configuration from a JSON file.
package config

import (
	"encoding/json"
	"fmt"
	"os"
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
	GrantMode     string `json:"grant_mode"` // "log" or "script"
	GrantCmd      string `json:"grant_cmd"`
	RevokeCmd     string `json:"revoke_cmd"`
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
