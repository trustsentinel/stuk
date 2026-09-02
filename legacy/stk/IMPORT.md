# Legacy import: stk

This directory is a clean-start snapshot (no history) of the `stk` prototype
— a secure lightweight broker for remote shell access (Noise Protocol + TOTP,
Go backend with a TypeScript/React terminal frontend).

It is imported as **reference code to modernize from**, not as the final
structure. The modern rewrite lives at the repository root (`pkg/`, `go.mod`).

Removed on import: thesis/report PDFs (Memoria/Manual/Resumen) and an example
TOTP secret in `auth/auth.go` (replaced with a placeholder).

Source: bluebycode/stk @ develop.
