#!/bin/sh
set -e
# --- sshd: demo user, key-based; keys come ONLY from stuk (time-gated) ---
ssh-keygen -A
adduser -D -s /bin/sh demo 2>/dev/null || true
# unlock the account so sshd accepts pubkey login (random throwaway pw; never used)
echo "demo:$(head -c 18 /dev/urandom | base64)" | chpasswd
sed -i 's/^#\?PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config

# The client's key is NOT placed in authorized_keys. Instead sshd asks stuk which
# keys are currently granted (AuthorizedKeysCommand), so a key works only during
# an active grant — the second gate, alongside the firewall.
mkdir -p /run/stuk/keys
cat >> /etc/ssh/sshd_config <<'EOF'
AuthorizedKeysCommand /usr/local/bin/stuk-authkeys -dir /run/stuk/keys
AuthorizedKeysCommandUser root
EOF

# --- firewall: gate tcp/22 (default DROP); knock/auth UDP stay open ---
# stukd's iptables backend inserts the per-client ACCEPT above this DROP on grant.
iptables -A INPUT -i lo -j ACCEPT
iptables -A INPUT -p tcp --dport 22 -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
iptables -A INPUT -p tcp --dport 22 -j DROP
echo "gateway: tcp/22 gated (default DROP), SSH keys gated via stuk-authkeys. starting sshd + stukd"

/usr/sbin/sshd
exec /usr/local/bin/stukd -config /app/stukd.json
