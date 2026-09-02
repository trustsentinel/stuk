#!/bin/sh
iptables -D INPUT -p tcp --dport 22 -s "$1" -j ACCEPT 2>/dev/null || true
echo "revoke.sh: closed tcp/22 for $1"
