#!/bin/sh
# End-to-end test: SSH blocked -> knock+TOTP -> SSH allowed -> TTL -> blocked again.
# Both gates are exercised: the firewall (iptables backend) AND the SSH key
# (authkeys backend via sshd AuthorizedKeysCommand). Exits non-zero on failure.
set -e
cd "$(dirname "$0")"
SECRET="JBSWY3DPEHPK3PXP"

cleanup() { docker compose down -v >/dev/null 2>&1 || true; }
trap cleanup EXIT

echo "== build & start =="
docker compose up -d --build
sleep 6

GW=$(docker compose exec -T gateway sh -c 'hostname -i' | tr -d '\r\n ')
echo "gateway IP: $GW"

ssh_try() {
  docker compose exec -T client sh -c \
    "ssh -i /keys/id -o StrictHostKeyChecking=no -o ConnectTimeout=5 -o BatchMode=yes demo@$GW 'echo SSH_OK' 2>/dev/null" || true
}

echo "== 1) SSH before knock (expect: blocked) =="
R1=$(ssh_try); echo "   -> '${R1:-<blocked>}'"

echo "== 2) knock sequence + TOTP (client sends its public key for the key gate) =="
docker compose exec -T client sh -c \
  "stuk -host $GW -ports 4000,4001,4002 -auth-port 4100 -secret $SECRET -pubkey \"\$(cat /keys/id.pub)\""
sleep 1

echo "== 3) SSH after knock (expect: SSH_OK) =="
R2=$(ssh_try); echo "   -> '${R2:-<blocked>}'"

echo "== 4) wait past TTL (22s), then SSH (expect: blocked) =="
sleep 22
R3=$(ssh_try); echo "   -> '${R3:-<blocked>}'"

echo
if [ -z "$R1" ] && [ "$R2" = "SSH_OK" ] && [ -z "$R3" ]; then
  echo "RESULT: PASS  (blocked -> granted -> auto-revoked; firewall + key both gated)"
  exit 0
fi
echo "RESULT: FAIL  R1='$R1' R2='$R2' R3='$R3'"
echo "--- gateway logs ---"; docker compose logs --no-log-prefix gateway | tail -25
exit 1
