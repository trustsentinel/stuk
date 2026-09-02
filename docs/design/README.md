# stuk — design docs

Architecture and design notes for the port-knocking SSH access manager,
translated from the original Spanish `stuk-docs` (2018–2019) and lightly edited.

- [Authentication](auth.md) — TOTP/2FA and key management
- [Agent](agent.md) — the port-knock server
- [Client](client.md) — the knock client and token format
- [Infrastructure](infrastructure.md) — reference deployment
- [Tools](tools.md) — build/runtime tooling
- [Troubleshooting](troubleshooting.md)

> These are historical design notes; the modern Go rewrite lives at the repo root.
> Lab credentials and internal addresses from the originals were removed.
