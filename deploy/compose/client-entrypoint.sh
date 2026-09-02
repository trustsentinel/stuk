#!/bin/sh
set -e
mkdir -p /keys
[ -f /keys/id ] || ssh-keygen -t ed25519 -N "" -f /keys/id
echo "client: key ready; idle."
exec sleep infinity
