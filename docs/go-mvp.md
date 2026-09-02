# stuk — Go MVP

A working port-knocking SSH access manager: `stukd` (daemon) watches for an
ordered UDP knock sequence, then grants temporary access to any source that
presents a valid TOTP code, auto-revoking after a TTL. `stuk` is the client.

## Build
```
go build ./cmd/stukd
go build ./cmd/stuk
```

## Run the daemon
```
cp examples/stukd.json stukd.json   # set totp_secret to a base32 secret
./stukd -config stukd.json
```
- `grant_mode: "log"` (default) just logs grants/revokes — safe for testing.
- `grant_mode: "script"` runs `grant_cmd` / `revoke_cmd` with `{ip}` and
  `{pubkey}` placeholders — e.g. an iptables rule or an `AuthorizedKeysCommand`
  update. This is where real SSH provisioning plugs in.

## Knock + authenticate (client)
```
./stuk -host SERVER -ports 4000,4001,4002 -auth-port 4100 \
       -secret <BASE32_TOTP_SECRET> -pubkey "ssh-ed25519 AAAA..."
```
The client sends the knock sequence, then an auth packet with a fresh TOTP code.
On success the daemon grants access for `ttl_seconds` and revokes automatically.

## Flow
1. Client knocks ports in order (UDP). Wrong port or a slow knock resets progress.
2. On the full sequence, the source is *armed* for `window_seconds`.
3. Client sends the TOTP code to `auth_port`. Valid code → temporary grant.
4. After `ttl_seconds`, access is revoked automatically.

> MVP scope: sequence detection is via UDP listeners (portable, no root/pcap).
> A future mode can use link-layer capture so knock ports stay fully closed
> (true stealth). TOTP uses `pkg/crypto` (pquerna/otp).
