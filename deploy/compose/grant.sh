#!/bin/sh
# Open SSH (tcp/22) for the client IP by inserting an ACCEPT above the DROP.
iptables -I INPUT -p tcp --dport 22 -s "$1" -j ACCEPT
echo "grant.sh: opened tcp/22 for $1"
