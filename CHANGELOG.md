# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-09-09
Initial tagged release.

### Added
- Go daemon (`stukd`) and client (`stuk`): port-knock + TOTP/2FA gated, time-limited SSH access.
- Pluggable grant backends: log, script, `iptables`, and an sshd AuthorizedKeysCommand (`stuk-authkeys`).
- Runnable Docker Compose end-to-end demo and GitHub Actions CI.

[0.1.0]: https://github.com/trustsentinel/stuk/releases/tag/v0.1.0
