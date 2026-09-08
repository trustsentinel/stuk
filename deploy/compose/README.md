# stuk — Docker Compose test scenario

A self-contained, runnable demo of the whole stuk flow:

> **SSH is blocked by default → a valid port-knock + TOTP opens it → access
> auto-revokes after the TTL.** Two gates are opened together: the **firewall**
> (`iptables` backend) and the **SSH key** (`authkeys` backend via sshd's
> `AuthorizedKeysCommand`), so the key works only while a grant is live.

It maps to the [architecture diagram](../../resources/stuk-architecture.png): a
client authenticates with a knock sequence + a time-based token, and the gateway
grants temporary SSH access.

## What's in the scenario

| Service | Role |
|---|---|
| **gateway** | Runs `stukd` **+** `sshd`, with `tcp/22` firewalled **closed** by default. A valid knock+TOTP makes stukd (`grant_mode: iptables,authkeys`) open `tcp/22` for the client IP **and** publish the client's key for `stuk-authkeys`; both are withdrawn after the TTL. Needs `NET_ADMIN`. |
| **client** | Runs the `stuk` client + an SSH client; generates a throwaway keypair into a shared volume and sends its public key in the knock. |

Only `tcp/22` is gated — the UDP knock ports (`4000-4002`) and auth port (`4100`)
stay reachable so knocks can arrive.

## Prerequisites
- Docker + Docker Compose v2
- The `gateway` needs `NET_ADMIN` (declared in `compose.yml`) to manage its firewall.

## Quick start — automated test
```bash
cd deploy/compose
./test-e2e.sh          # exits non-zero on failure; tears down on exit
```
It builds the images, starts both services, and asserts:
1. SSH **before** knocking is blocked
2. knock sequence + TOTP sent (with the client's public key)
3. SSH **after** knocking returns `SSH_OK`
4. after the TTL (20s) SSH is blocked again

Expected final line: `RESULT: PASS  (blocked -> granted -> auto-revoked; firewall + key both gated)`

## Manual walk-through
```bash
cd deploy/compose
docker compose up -d --build
GW=$(docker compose exec -T gateway hostname -i | tr -d '\r\n ')

# 1) blocked (times out)
docker compose exec client ssh -i /keys/id -o ConnectTimeout=5 -o BatchMode=yes demo@$GW 'echo hi'

# 2) knock + authenticate, sending the client's public key for the key gate
docker compose exec client sh -c \
  "stuk -host $GW -ports 4000,4001,4002 -auth-port 4100 -secret JBSWY3DPEHPK3PXP -pubkey \"\$(cat /keys/id.pub)\""

# 3) SSH now works (within the TTL)
docker compose exec client ssh -i /keys/id -o StrictHostKeyChecking=no demo@$GW 'echo SSH_OK'

# 4) after ~20s the grant expires: firewall closes AND the key is withdrawn
docker compose logs gateway   # shows: sequence complete -> ACCESS GRANTED -> (revoke)
```

## Configuration
Edit [`stukd.json`](stukd.json): `knock_ports` / `auth_port`, `window_seconds`,
`ttl_seconds` (demo 20s), `totp_secret` (fixed test value), and
`grant_mode` (`iptables,authkeys` here). See the repo README for all backends.

## Cleanup
```bash
docker compose down -v
```

## Demo simplifications (vs production)
- **UDP listeners** detect knocks, so the knock ports are technically open.
  Production should use link-layer capture (pcap/AF_PACKET) for true stealth.
- The gateway doubles as **broker + SSH target**; production separates them.
- The TOTP secret is a **fixed test value** and the demo user is unlocked with a
  **random throwaway password** (pubkey-only) — never do either in production.
- Prefer **short-lived SSH certificates** (SSH CA) over authorized-key gating for
  larger fleets.
