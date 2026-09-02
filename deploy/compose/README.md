# stuk — Docker Compose test scenario

A self-contained, runnable demo of the whole stuk flow:

> **SSH is blocked by default → a valid port-knock + TOTP opens it → access
> auto-revokes after the TTL.**

It maps directly to the project's [architecture diagram](../../resources/stuk-architecture.png):
a client authenticates with a knock sequence + a time-based token, and the
gateway grants temporary SSH access.

## What's in the scenario

| Service | Role |
|---|---|
| **gateway** | Runs `stukd` (the daemon) **+** `sshd`, with `tcp/22` firewalled **closed** by default. A successful knock+TOTP triggers `grant.sh`, which opens `tcp/22` for the client's IP; `revoke.sh` closes it again after the TTL. Needs `NET_ADMIN`. |
| **client** | Runs the `stuk` client and an SSH client. Generates a throwaway SSH keypair into a shared volume. |

Only `tcp/22` is gated — the UDP knock ports (`4000-4002`) and auth port (`4100`)
stay reachable so knocks can arrive.

## Prerequisites
- Docker + Docker Compose v2
- The `gateway` needs the `NET_ADMIN` capability (declared in `compose.yml`) to manage its firewall.

## Quick start — automated test
```bash
cd deploy/compose
./test-e2e.sh
```
It builds the images, starts both services, and asserts:
1. SSH **before** knocking is blocked
2. knock sequence + TOTP is sent
3. SSH **after** knocking returns `SSH_OK`
4. after the TTL (20s) SSH is blocked again

Expected final line: `RESULT: PASS ✅  (blocked -> granted -> auto-revoked)`

## Manual walk-through
```bash
cd deploy/compose
docker compose up -d --build

# gateway IP on the compose network
GW=$(docker compose exec -T gateway hostname -i | tr -d '\r\n ')

# 1) SSH is blocked (times out)
docker compose exec client ssh -i /keys/id -o ConnectTimeout=5 -o BatchMode=yes demo@$GW 'echo hi'

# 2) knock + authenticate (client derives the TOTP code from the shared demo secret)
docker compose exec client stuk -host $GW -ports 4000,4001,4002 -auth-port 4100 -secret JBSWY3DPEHPK3PXP

# 3) SSH now works (within the TTL)
docker compose exec client ssh -i /keys/id -o StrictHostKeyChecking=no demo@$GW 'echo SSH_OK'

# 4) after ~20s the grant expires and SSH is blocked again
docker compose logs gateway   # shows: sequence complete -> ACCESS GRANTED -> (revoke)
```

## Configuration
Edit [`stukd.json`](stukd.json):
- `knock_ports` / `auth_port` — the sequence and auth port
- `window_seconds` — max time to complete the sequence
- `ttl_seconds` — how long access stays open (demo: 20s)
- `totp_secret` — base32 secret (demo uses a fixed test value; the client passes the same via `-secret`)
- `grant_mode` — `script` here (runs `grant.sh`/`revoke.sh`); `log` for a dry run

## Cleanup
```bash
docker compose down -v
```

## Demo simplifications (vs production)
- **UDP listeners** are used for knock detection, so the knock ports are technically open. Production should use link-layer capture (pcap/AF_PACKET) so they stay fully closed — true stealth.
- The gateway doubles as **supervisor + SSH target** for simplicity; production separates them, and the supervisor provisions access on remote hosts.
- The TOTP secret is a **fixed test value** and the demo user is unlocked with a **random throwaway password** (pubkey-only login) — never do either in production.
- Real deployments should provision access via **short-lived SSH certificates** (SSH CA) rather than authorized_keys + firewall rules.
