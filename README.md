# Stuk — Secure SSH Access Manager

[![CI](https://github.com/trustsentinel/stuk/actions/workflows/ci.yml/badge.svg)](https://github.com/trustsentinel/stuk/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

<p align="center">
  <a href="https://www.incibe.es/"><img src="docs/assets/img/incibe.png" alt="INCIBE — Instituto Nacional de Ciberseguridad" height="64"></a>
</p>
<p align="center">
  <sub>🏆 <b>INCIBE National Cybersecurity Competition 2018</b> — stealth SSH via port-knocking.</sub>
</p>

SSH access management using **port knocking** and **TOTP/2FA**. SSH stays closed
until a client sends the right knock sequence and a valid time-based code — then
access is granted temporarily and revoked automatically.

## Features

- 🔐 **Port knocking** — a secret knock sequence before SSH is reachable
- 🔑 **TOTP/2FA** — time-based one-time-password verification
- ⏱️ **Temporary access** — grants expire and auto-revoke after a TTL
- 🧩 **Pluggable grants** — log (dry run) or script (iptables / AuthorizedKeysCommand)
- 🚀 **Single static binary** — daemon (`stukd`) + client (`stuk`)

## Architecture

A client obtains a time-based token (e.g. Google Authenticator), then knocks with
`stuk`. The **daemon** (`stukd`) verifies the TOTP code and provisions temporary
access to the target servers — access that expires automatically.

```mermaid
flowchart LR
  client["client<br/>(stuk)"]
  daemon["stukd<br/>verifies knock + TOTP · grants TTL access"]
  ssh["SSH servers<br/>(closed by default)"]

  client -->|"knock sequence (UDP)"| daemon
  client -->|"TOTP token"| daemon
  daemon -->|"grant, auto-revoked after TTL"| ssh
```

## Quick start

### Install
```bash
go install github.com/trustsentinel/stuk/cmd/stukd@latest   # daemon
go install github.com/trustsentinel/stuk/cmd/stuk@latest    # client
```

### Run the daemon
```bash
cp examples/stukd.json stukd.json      # set totp_secret to a base32 secret
stukd -config stukd.json
```

### Knock + authenticate (client)
```bash
stuk -host SERVER -ports 4000,4001,4002 -auth-port 4100 \
     -secret YOUR_BASE32_TOTP_SECRET -pubkey "ssh-ed25519 AAAA..."
```
The client sends the knock sequence, then a TOTP-authenticated request. On
success the daemon grants access for the configured TTL and revokes it
automatically. Full guide: [docs/go-mvp.md](docs/go-mvp.md).

## Try it end-to-end (Docker)

A runnable demo — a gated `sshd` gateway plus `stukd`, with SSH firewalled closed
by default — lives in [`deploy/compose/`](deploy/compose):

```bash
cd deploy/compose && ./test-e2e.sh
```
It proves the whole flow: **SSH blocked → knock + TOTP → SSH allowed → auto-revoked.**

## Configuration

The daemon reads a JSON config (see [`examples/stukd.json`](examples/stukd.json)):

| Key | Meaning |
|---|---|
| `knock_ports` | ordered UDP knock sequence |
| `auth_port` | port that receives the TOTP auth packet |
| `window_seconds` | max time to complete the sequence |
| `ttl_seconds` | how long access stays open |
| `totp_secret` | base32 TOTP secret |
| `grant_mode` | comma-separated backends (below): `log`, `script`, `iptables`, `authkeys` |
| `ssh_port` / `iptables_chain` | `iptables` backend: port to open (default 22) and chain (default `INPUT`) |
| `authkeys_dir` | `authkeys` backend: directory of granted keys (default `/run/stuk/keys`) |

### Grant backends
On a valid knock+TOTP, stukd provisions access for `ttl_seconds` then auto-revokes.
`grant_mode` selects one or more backends (applied together):

- **`log`** — dry run; logs only (default).
- **`script`** — runs `grant_cmd` / `revoke_cmd` with `{ip}` / `{pubkey}` substituted.
- **`iptables`** — natively inserts `-I <chain> -p tcp --dport <ssh_port> -s <ip> -j ACCEPT`
  on grant and deletes it on revoke (needs `NET_ADMIN`). The firewall gate.
- **`authkeys`** — writes the client's public key into `authkeys_dir` for the grant's
  lifetime. Point sshd at [`stuk-authkeys`](cmd/stuk-authkeys) so the key works
  *only* while granted — the key gate, on top of the firewall:
  ```
  # /etc/ssh/sshd_config
  AuthorizedKeysCommand      /usr/local/bin/stuk-authkeys -dir /run/stuk/keys
  AuthorizedKeysCommandUser  root
  ```

Combine them, e.g. `"grant_mode": "iptables,authkeys"` — the runnable demo in
[`deploy/compose/`](deploy/compose) uses exactly this (network **and** key gated).

## Development
```bash
git clone https://github.com/trustsentinel/stuk.git && cd stuk
go build ./cmd/stukd ./cmd/stuk
go test ./...
```

### Project structure
```
stuk/
├── cmd/
│   ├── stuk/           # client: sends the knock sequence + TOTP
│   ├── stukd/          # daemon: detects knocks, verifies TOTP, grants access
│   └── stuk-authkeys/  # sshd AuthorizedKeysCommand: prints currently-granted keys
├── internal/
│   ├── knock/          # ordered knock-sequence detection + senders
│   ├── grant/          # backends (log/script/iptables/authkeys) + TTL auto-revoke
│   └── config/         # daemon JSON config
├── pkg/crypto/         # TOTP (pquerna/otp)
├── deploy/compose/     # runnable end-to-end Docker demo
├── docs/               # design notes (docs/design) + Go MVP guide
└── examples/           # example stukd.json
```

## How it works
1. Client knocks the ports in order (UDP). A wrong port or a slow knock resets progress.
2. On the full sequence, the source is *armed* for `window_seconds`.
3. Client sends the TOTP code to `auth_port`; a valid code → temporary grant.
4. After `ttl_seconds`, access is revoked automatically.

## Security notes
- Use a strong base32 TOTP secret; never commit real secrets.
- Keep knock ports off the public internet where possible.
- Prefer short TTLs (5–15 min) and short-lived SSH certificates over long-lived keys.
- **Roadmap:** link-layer capture so knock ports stay fully closed (true stealth),
  encrypted auth channel, per-user secrets, and SSH-CA short-lived certificates —
  see [TASKS](https://github.com/trustsentinel) / `docs/`.

## TrustSentinel
Part of [TrustSentinel](https://trustsentinel.eu) — secure connectivity and
network-intelligence tooling by Álvaro López.

- **[netso](https://github.com/trustsentinel/netso)** — secure-networking platform (SSI + end-to-end encryption)
- **[stk](https://github.com/trustsentinel/stk)** — browser-based remote shell broker
- **[stuk](https://github.com/trustsentinel/stuk)** — SSH access gating (port-knock + MFA)  ·  _this repo_
- **[argos](https://github.com/trustsentinel/argos)** — P2P blockchain network scanning
- **[eth-rlp](https://github.com/trustsentinel/eth-rlp)** — RLP codec for Ethereum discv4

## License
MIT — see [LICENSE](LICENSE).

## Acknowledgments
Originally prototyped by [@bluebycode](https://github.com/bluebycode) (INCIBE 2018,
stealth SSH via port-knocking); rewritten in Go under TrustSentinel. Original design
notes are translated in [`docs/design/`](docs/design).
