#!/bin/sh
set -e
# --- sshd: demo user, key-based, from the shared /keys volume ---
ssh-keygen -A
adduser -D -s /bin/sh demo 2>/dev/null || true
# unlock the account so sshd accepts pubkey login (random throwaway pw; never used)
echo "demo:$(head -c 18 /dev/urandom | base64)" | chpasswd
mkdir -p /home/demo/.ssh
echo "gateway: waiting for client public key on /keys/id.pub ..."
for i in $(seq 1 60); do [ -f /keys/id.pub ] && break; sleep 1; done
cat /keys/id.pub > /home/demo/.ssh/authorized_keys 2>/dev/null || { echo "no client key"; exit 1; }
chown -R demo:demo /home/demo/.ssh && chmod 700 /home/demo/.ssh && chmod 600 /home/demo/.ssh/authorized_keys
sed -i 's/^#\?PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config

# --- firewall: gate tcp/22 (default DROP); knock/auth UDP stay open ---
iptables -A INPUT -i lo -j ACCEPT
iptables -A INPUT -p tcp --dport 22 -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
iptables -A INPUT -p tcp --dport 22 -j DROP
echo "gateway: tcp/22 gated (default DROP). starting sshd + stukd"

/usr/sbin/sshd
exec /usr/local/bin/stukd -config /app/stukd.json
