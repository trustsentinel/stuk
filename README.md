# Stuk - Secure SSH Access Manager

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Modern Go implementation of SSH access management using port knocking and MFA/TOTP authentication.

## Features

- 🔐 **Port Knocking** - Secure knock sequence before SSH access
- 🔑 **TOTP/2FA** - Time-based one-time password authentication
- ⏱️ **Temporary Access** - Automatic SSH key provisioning with expiration
- 🚀 **Zero Dependencies** - Single binary deployment
- 📦 **Docker Ready** - Containerized deployment
- 🔒 **Security First** - Type-safe Go implementation

## Architecture

```
┌─────────┐     Knock      ┌────────────┐     Auth      ┌─────────────┐
│ Client  │ ─────────────> │ Supervisor │ ───────────> │ SSH Servers │
└─────────┘   TOTP Token   └────────────┘    Grant      └─────────────┘
                                              Access
```

## Quick Start

### Installation

```bash
go install github.com/trustsentinel/stuk/cmd/client@latest
go install github.com/trustsentinel/stuk/cmd/supervisor@latest
```

### Using Docker

```bash
docker pull trustsentinel/stuk:latest

# Run client
docker run trustsentinel/stuk:latest stuk-client \
  --token 123456 \
  --secret YOUR_SECRET \
  --host server.example.com

# Run supervisor
docker run -d \
  -p 7000-9000:7000-9000 \
  trustsentinel/stuk:latest stuk-supervisor
```

## Usage

### Client

Generate TOTP token using Google Authenticator or similar app, then:

```bash
stuk-client \
  --token 123456 \
  --secret YOUR_TOTP_SECRET \
  --host target.example.com \
  --ports 7000,8000,9000
```

Or using environment variables:

```bash
export STUK_SECRET=YOUR_TOTP_SECRET
stuk-client --token 123456 --host target.example.com
```

### Supervisor

Run the supervisor daemon on your infrastructure:

```bash
stuk-supervisor \
  --listen 0.0.0.0:7000 \
  --ports 7000,8000,9000 \
  --keydir /etc/stuk/keys \
  --duration 5m
```

## Configuration

### Port Knocking Sequence

Default sequence: `7000,8000,9000`

Customize with `--ports` flag:

```bash
--ports 1234,5678,9012
```

### Access Duration

Default: 5 minutes

Customize with `--duration` flag:

```bash
--duration 10m
```

## Development

### Build from Source

```bash
git clone https://github.com/trustsentinel/stuk.git
cd stuk

# Build client
go build -o stuk-client ./cmd/client

# Build supervisor
go build -o stuk-supervisor ./cmd/supervisor
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./pkg/knock
```

### Project Structure

```
stuk/
├── cmd/
│   ├── client/         # CLI client application
│   └── supervisor/     # Supervisor daemon
├── pkg/
│   ├── crypto/         # TOTP/cryptography
│   ├── knock/          # Port knocking logic
│   └── ssh/            # SSH key management
├── Dockerfile
├── go.mod
└── README.md
```

## Security Considerations

- 🔒 Always use strong TOTP secrets (minimum 160 bits)
- 🔐 Enable firewall rules to only allow supervisor
- ⏱️ Set appropriate access durations (5-15 minutes recommended)
- 🔑 Rotate SSH keys regularly
- 📝 Monitor supervisor logs for suspicious activity
- 🚫 Never expose knock ports to public internet without firewall

## How It Works

1. **Setup**: User enrolls with TOTP (Google Authenticator)
2. **Knock**: Client sends knock sequence to supervisor
3. **Validate**: Supervisor validates TOTP token
4. **Grant**: Supervisor temporarily adds SSH key to authorized_keys
5. **Access**: User connects via SSH within time window
6. **Expire**: SSH key automatically removed after duration

## Performance

- **Knock latency**: < 500ms for 3-port sequence
- **TOTP validation**: < 1ms
- **Memory footprint**: ~8MB (client), ~15MB (supervisor)
- **Binary size**: ~8MB (static compilation)

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass (`go test ./...`)
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file

## Acknowledgments

Original Python implementation by [@vrandkode](https://github.com/vrandkode)

Modernized and rewritten in Go for improved security and performance.

## Support

- 🐛 Issues: [GitHub Issues](https://github.com/trustsentinel/stuk/issues)
- 📧 Email: github.com/trustsentinel
- 💬 Discussions: [GitHub Discussions](https://github.com/trustsentinel/stuk/discussions)
