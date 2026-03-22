# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Tern is a WebSocket relay server (Go) that enables connections between
devices where the backend is on a private network with no ingress. It
establishes a virtual WebSocket connection over an actual WebSocket
connection such that the relay itself cannot inspect the traffic.

Extracted from the jevon project. Deployed on Fly.io (Sydney region).

## Build, Test & Run

```bash
go build -o tern .           # build
go test ./...                # all tests (relay + crypto)
go test ./crypto/            # crypto tests only
./tern                       # run (listens on :8080)
PORT=3000 ./tern             # custom port
```

## Architecture

### Relay (`main.go`)

Bidirectional WebSocket forwarder. Backends register and get an instance
ID; clients connect by ID and traffic is bridged opaquely.

**Core types:**
- `relay` — thread-safe registry of active backend instances (RWMutex-protected map)
- `instance` — a registered backend: ID, WebSocket conn, context, write mutex

**Endpoints:**
- `GET /health` — JSON health check
- `GET /register` — backend connects via WebSocket, receives a random base36 ID
- `GET /ws/{id}` — client connects via WebSocket, bidirectional bridge to backend

### E2E Crypto (`crypto/`)

Application-level encryption so the relay sees only ciphertext:
- **Key exchange:** X25519 ECDH
- **Symmetric encryption:** AES-256-GCM with monotonic counter nonce
- **Key derivation:** HKDF-SHA256
- **Confirmation codes:** 6-digit order-independent code from both ECDH pubkeys (MitM detection)

### Pairing Protocol (`protocol/pairing.yaml`)

Defines the pairing ceremony: QR code distribution → ECDH key exchange
through the relay → confirmation code verification → encrypted device
secret delivery → nonce-based session authentication on reconnect.

### Formal Model (`formal/`)

TLA+ specification (`PairingCeremony.tla`) with adversary model
verifying: no token reuse, MitM detection via code mismatch, device
secret secrecy, auth requires completed pairing, no nonce reuse.

Run with `./formal/tlc PairingCeremony`.

## Deployment

Fly.io config in `fly.toml`. Multi-stage Dockerfile (`golang:1.25-alpine`
→ `alpine:3.21`). Shared-cpu-1x, 256MB, auto-start/stop with zero
minimum machines.

## Delivery

Merged to master.
