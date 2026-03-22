# Convergence Targets

## 🎯T1 Tern is a complete library for opaque authenticated relay

All crypto, protocol state machines, code generators, QR helper, and
Swift package live here. Applications import tern rather than duplicating
relay/pairing logic.

**Sub-targets:**

### 🎯T1.1 Crypto library migrated from jevon

`crypto/` package with key exchange, symmetric encryption, confirmation
code derivation, and all tests. No jevon-specific imports.

Status: done

### 🎯T1.2 Pairing protocol spec migrated from jevon

`protocol/pairing.yaml` lives here as the source of truth.

Status: done

### 🎯T1.3 TLA+ formal model migrated from jevon

`formal/PairingCeremony.tla`, configs, and `tlc` wrapper script.
Ephemeral state files and trace files excluded.

Status: done

### 🎯T1.4 Protocol state machine framework migrated from jevon

`protocol/` package: declarative state machine framework with YAML
parser, runtime executor, and code generators (Go, Swift, TLA+,
PlantUML). `cmd/protogen/` generates all outputs from YAML spec.

Status: done

### 🎯T1.5 QR helper migrated from jevon

`qr/` package: terminal QR code rendering and LAN IP detection.
Jevon-specific URL scheme removed.

Status: done

### 🎯T1.6 Swift package (SPM)

`Package.swift` at repo root, `Sources/TernCrypto/` with E2ECrypto.swift
and generated PairingCeremonyMachine.swift. All types public. Tests pass.

Status: done

### 🎯T1.7 E2E integration test

Full-stack test exercising relay + ECDH pairing + confirmation codes +
encrypted bidirectional messaging. Verifies relay only sees ciphertext.

Status: done

### 🎯T1.8 Jevon imports tern's packages

Jevon's `internal/crypto/`, `internal/protocol/`, `internal/qr/`, and
`cmd/protogen/` are replaced by imports from tern. iOS app imports
TernCrypto SPM package. This validates the extraction is clean.

Status: not started (requires tern to be tagged and pushed)
