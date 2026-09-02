# Authentication

## Goals

**Authentication**
1. Bind a user (U) under a domain to an authenticator app, generating a QR code (Q). From there the authenticator provides temporary secret keys via [**TOTP**](https://tools.ietf.org/html/rfc6238) (Sk).
2. Two possible strategies:
   - **a.** Store a generated key (Bk) per user (U), used as a [**salt**](https://en.wikipedia.org/wiki/Salt_(cryptography)).
   - **b.** Otherwise, no per-user key persistence — a single salt key for everyone (Ak).
3. Verify the TOTP key through an API. Generally not exposed publicly — only needed by the infrastructure.

**Key management** (if strategy 2.a)
4. Persist users with their keys: `U => Bk`.

## Functional requirements

**Authentication**
- Endpoint offering secure verification of a key (REST).
- Web service that generates a QR code.

**Key management**
- Service to register and retrieve `U => Bk` relations (NoSQL, e.g. [Redis](https://redis.io/)).

## Possible improvements
- **Secure server** — domain + HTTPS (e.g. certbot) on the public-facing service (highly desirable).
- **Any 2FA-TOTP app** — integrate with any TOTP 2FA implementation (Google Authenticator, etc.), including docs, or a custom TOTP 2FA implementation on Android or another platform.

## References
- RFC 6238 — TOTP: Time-Based One-Time Password Algorithm.
