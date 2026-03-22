# Convergence Targets

## 🎯T1 E2E encryption and pairing ceremony are part of tern

The crypto library (X25519 ECDH, AES-256-GCM, HKDF key derivation),
pairing ceremony protocol spec, and TLA+ formal model live in this
repo — not jevon. Tern's value proposition is opaque relay with
authenticated pairing, not just dumb forwarding.

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

### 🎯T1.4 Jevon imports tern's crypto package

Jevon's `internal/crypto/` is replaced by an import of
`github.com/marcelocantos/tern/crypto`. This validates that the
extraction is clean.

Status: not started (depends on 🎯T1.1)
