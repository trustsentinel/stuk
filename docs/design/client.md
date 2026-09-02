# Client

## Features
- SSH functional wrapper *(incomplete)*.
- Generates a public/private key pair for the key exchange *{Upub, Upri}*.
- Sends TCP sequences with a token formed by:

  ```
  Token ::= Ip, ENC(Ipub, Upub), HMAC(Sk)
  ```
  - **Ipub:** infrastructure public key
  - **Upub, Upriv:** volatile key pair
  - **Sk:** secret key (one-time password) sent by the authenticator app

- Generate an SSH RSA private/public key pair in Go.

## References
- A port-knocking implementation: `jvinet/knock` (client) / zeroflux `knock` (agent).
- Python cryptography libraries.
- Go examples for the client wrapper.
