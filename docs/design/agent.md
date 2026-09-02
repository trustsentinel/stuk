# Agent

**Role: port-knocking server.**

It is a port-knock server. It listens to all traffic on an Ethernet interface
inside an infrastructure's private network, looking for special sequences of
port "hits". A client performs these port visits by sending a TCP (or UDP)
packet to a port on one of the end systems (or the same system). That port does
not need to be open — the agent listens at the link layer and sees all traffic,
even for closed ports. When the server detects a specific sequence of port
visits, it runs a command defined in its config file. This can open holes in a
firewall for quick access.

Reference: Port Knocking — Krzywinski, *SysAdmin* 2003.

## Features

**Packet filtering.** Listen on the supervisor's interface for the port-knock
sequences redirected by the end machines. Intercepted packets are parsed to obtain:
- end-machine IP
- client IP
- token
- sequence

The agent only filters packets matching a BPF expression such as:
```
ip proto \tcp and ((tcp dst port 4000 or 4001 or 4002) and tcp[tcpflags] & tcp-syn != 0)
```

**Token verification.** Verify the obtained token; a callback runs some
functionality on the system after verification.

**Public-key handling** on the end systems, distributed. Just-in-time SSH
provisioning via `AuthorizedKeysCommand`:
```
# /etc/ssh/sshd_config
AuthorizedKeysCommand /path/to/iam-ssh-auth
AuthorizedKeysCommandUser nobody
```
- Expire public keys that are no longer valid.

## Non-functional requirements
- Virtualized infrastructure as a proof of concept (Docker, VMware, or VirtualBox).
- Redirecting traffic from one machine to another (iptables).
