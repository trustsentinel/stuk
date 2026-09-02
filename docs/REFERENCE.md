# stuk — reference & clean-room policy

`trustsentinel/stuk` is a **clean-room Go rewrite** of a port-knocking SSH access
manager. It is NOT a code import.

## Original reference (read-only)

- **[bluebycode/stuk](https://github.com/bluebycode/stuk)** (2019, Ruby/Python) —
  port-knocking SSH access manager with automatic SSH key provisioning and
  MFA/TOTP + U2F. Original hackathon prototype (CyberCamp 2019).

We treat it as **read-only reference**: its design and logic inform this rewrite,
but its git history is **deliberately not imported**.

## Why clean-room (not imported)

`bluebycode/stuk` has been **public since 2018** and its history contains committed
secrets — a Rails `master.key`, a Devise signing secret, and private keys including
`infrastructure/keys/cybercamp.pem`. Those are considered **compromised** (long
public) and must never enter this repository. A full-history `gitleaks` scan is
recorded in the migration's scan reports.

Importing-then-scrubbing would drag that history along; a clean-room rewrite avoids
it entirely.

## What to carry over (concepts, re-implemented in Go)

- Port-knock sequence design and the knock-daemon → SSH-grant flow
- TOTP/2FA enrollment + verification (root scaffold already has `pkg/crypto/totp.go`)
- Temporary SSH key provisioning with expiry
- Supervisor/agent model for managing access on remote hosts

## What to modernize (see MIGRATION.md Phase 5)

- U2F → WebAuthn / passkeys
- Keys → ed25519, hardware-backed where available
- GitHub Actions CI with `govulncheck` + `gosec`
