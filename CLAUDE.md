# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Tern is a WebSocket relay library and server (Go + Swift) that enables
connections between devices where the backend is on a private network
with no ingress. It establishes a virtual WebSocket connection over an
actual WebSocket connection such that the relay itself cannot inspect
the traffic.

Applications import tern's packages rather than implementing
relay/pairing/crypto logic themselves.

## Build, Test & Run

```bash
go build -o tern .                # build relay server
go test ./...                     # all Go tests (relay, crypto, protocol, E2E)
go test -run TestE2E              # E2E integration test only
swift test                        # Swift crypto + state machine tests
JAVA_HOME=<jdk21> android/gradlew -p android test  # Kotlin tests
go run ./cmd/protogen protocol/pairing.yaml  # regenerate from YAML spec
```

## Package Structure

### Relay Server (`main.go`)

Bidirectional WebSocket forwarder. Backends register and get an instance
ID; clients connect by ID and traffic is bridged opaquely.

**Endpoints:** `GET /health`, `GET /register`, `GET /ws/{id}`

### E2E Crypto (`crypto/`)

Application-level encryption so the relay sees only ciphertext:
- **Key exchange:** X25519 ECDH
- **Symmetric encryption:** AES-256-GCM with monotonic counter nonce
- **Key derivation:** HKDF-SHA256
- **Confirmation codes:** 6-digit order-independent code from both ECDH pubkeys (MitM detection)

### Protocol Framework (`protocol/`)

Declarative state machine framework. Protocols are defined in YAML
(`protocol/pairing.yaml`) and serve as the single source of truth for:
- Go runtime executor (`machine.go`)
- Go code generator (`gogen.go`)
- Swift code generator (`swift.go`)
- TLA+ spec generator (`tla.go`)
- PlantUML diagram generator (`plantuml.go`)

### Code Generator (`cmd/protogen/`)

Reads YAML, validates, and generates Go/Swift/TLA+/PlantUML outputs.

### QR Helper (`qr/`)

Terminal QR code rendering and LAN IP detection for device pairing flows.

### Formal Model (`formal/`)

TLA+ specification (`PairingCeremony.tla`) with adversary model verifying:
no token reuse, MitM detection via code mismatch, device secret secrecy,
auth requires completed pairing, no nonce reuse.

Run with `./formal/tlc PairingCeremony`.

### Swift Package (`Package.swift`, `Sources/TernCrypto/`)

SPM library (`TernCrypto`) containing E2ECrypto.swift and the generated
PairingCeremonyMachine.swift. iOS apps add the GitHub repo as a package
dependency.

### Android/Kotlin Library (`android/terncrypto/`)

Kotlin/JVM library mirroring TernCrypto: `E2EKeyPair`, `E2EChannel`,
`Hkdf`, and the generated `PairingCeremonyMachine.kt`. Consumed via
JitPack (`com.github.marcelocantos.tern:terncrypto:<tag>`).
Requires JDK 17+ / Android API 33+ (for X25519 support).

## Deployment

Fly.io config in `fly.toml`. Multi-stage Dockerfile (`golang:1.25-alpine`
→ `alpine:3.21`). Shared-cpu-1x, 256MB, auto-start/stop with zero
minimum machines.

## Delivery

Merged to master.
