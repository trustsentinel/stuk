# stuk — origin, contents & modernization

`trustsentinel/stuk` is a **port-knocking SSH access manager**. This repository now
contains the **sanitized original implementation** imported from
[bluebycode/stuk](https://github.com/bluebycode/stuk) (2019), plus a Go rewrite
scaffold that will grow alongside it.

## Layout

- `client/` — Python port-knocking client (`pnock.py`, `stuk.py`, `crypto.py`)
- `supervisor/` — Python supervisor daemon (knock listener → SSH provisioning)
- `auth/` — Ruby on Rails auth / dashboard server
- `documentacion/` — original docs (Spanish; translate → English, see MIGRATION Phase 4)
- `docs/legacy-README.md` — the original project README
- `pkg/crypto/totp.go`, `go.mod` — start of the Go rewrite (`cmd/` to be added, Phase 5)

## Sanitization on import (2026-09-02)

The original was **public since 2018** and contained committed secrets. On import,
ALL of the following were removed / scrubbed, and a full `gitleaks` scan is clean:

- Removed: `auth/config/master.key`, `auth/config/credentials.yml.enc`,
  `auth/certs/localhost.{key,crt}`, `client/tests/.keys/*.pem`,
  `infrastructure/keys/cybercamp.pem`
- Scrubbed: commented Devise `secret_key` in `auth/config/initializers/devise.rb`;
  PII in a `supervisor/supervisor.py` docstring (real email + SSH public key + internal IP)
- Removed: hackathon presentation PDFs

> ⚠️ Those original keys have been public since 2018 — treat as **compromised**.
> Rotate/revoke anything still live (esp. the CyberCamp deploy key) on the source side.

## Modernization targets (Phase 5)

- Add `cmd/stuk` entrypoint; port the knock→grant flow to Go
- TOTP/2FA (scaffold has `pkg/crypto/totp.go`); U2F → WebAuthn/passkeys
- ed25519, hardware-backed keys; GitHub Actions CI + `govulncheck` + `gosec`
- Translate `documentacion/` ES→EN
