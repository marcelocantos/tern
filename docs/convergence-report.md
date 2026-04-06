# Convergence Report

Evaluated: 2026-04-06

## Standing invariants

- **Tests**: PASSING (CI run #24020825119 succeeded)
- **CI**: GREEN — all jobs passing on master

## Movement

- 🎯T19: not started -> **achieved** (hierarchical state machines landed in #3)
- 🎯T5.1: not started -> **achieved** (was already done, now confirmed and moved to Achieved)
- Standing invariant: FAILING -> **GREEN** (Dockerfile fixed in #2, `COPY protocol/` and `COPY qr/` added)
- 🎯T1, 🎯T1.8, 🎯T5, 🎯T6, 🎯T7, 🎯T8, 🎯T10, 🎯T14, 🎯T15, 🎯T17: (unchanged)

## Gap Report

### 🎯T1.8 Jevon imports tern's packages  [weight: 1.7]
Gap: not started
v0.9.0 released. Migration in jevon repo still pending. This is external work in the jevon codebase — requires tagging tern and updating jevon's imports.

### 🎯T6 Investigate STUN/NAT hole-punching  [weight: 1.5]
Gap: not started
Pure research target. No code or investigation artifacts exist yet.

### 🎯T17 Makefile deploy target  [weight: 1.5]
Gap: not started
No Makefile deploy target exists. CI deploy is working now (standing invariant fixed), so this is unblocked infrastructure work.

### 🎯T1 Tern is a complete library  [weight: 1.7]
Gap: converging (7/8 sub-targets achieved)

  [x] 🎯T1.1 Crypto library — achieved
  [x] 🎯T1.2 Pairing protocol spec — achieved
  [x] 🎯T1.3 TLA+ formal model — achieved
  [x] 🎯T1.4 Protocol state machine framework — achieved
  [x] 🎯T1.5 QR helper — achieved
  [x] 🎯T1.6 Swift package — achieved
  [x] 🎯T1.7 E2E integration test — achieved
  [ ] 🎯T1.8 Jevon imports tern's packages — not started

### 🎯T8 WebTransport relay  [weight: 1.0]
Gap: converging (3/5 sub-targets achieved)

  [x] 🎯T8.1 WebTransport relay server — achieved
  [x] 🎯T8.2 Non-strict Channel.Decrypt — achieved
  [x] 🎯T8.3 Go WebTransport client — achieved
  [ ] 🎯T8.4 Web/TypeScript WebTransport client — not started
  [ ] 🎯T8.5 LAN direct WebTransport — not started (blocked on 🎯T8.4)

### 🎯T14 Browser WebTransport E2E  [weight: 1.0]  (status only)
Status: blocked on Playwright headless Chromium QUIC support

### 🎯T15 Gomobile bindings  [weight: 1.0]  (status only)
Status: not started

### 🎯T7 Investigate Bluetooth as proximity oracle  [weight: 1.0]  (status only)
Status: not started

### 🎯T5 Multi-transport with LAN upgrade  [weight: 0.8]
Gap: converging (1/4 sub-targets achieved)

  [x] 🎯T5.1 Reorder-tolerant decryption — achieved
  [ ] 🎯T5.2 LAN discovery via relay — not started
  [ ] 🎯T5.3 Cutover protocol — not started (blocked on 🎯T10)
  [ ] 🎯T5.4 Transport-agnostic Conn — not started (blocked on 🎯T5.2, 🎯T5.3)

### 🎯T10 TLA+ model for cutover protocol  [weight: 0.6]  (status only)
Status: not started. Low effective weight — blocks 🎯T5.3 but expensive relative to value.

### Blocked targets

- 🎯T5.3 Cutover protocol — blocked on 🎯T10
- 🎯T5.4 Transport-agnostic Conn — blocked on 🎯T5.2, 🎯T5.3
- 🎯T8.5 LAN direct WebTransport — blocked on 🎯T8.4

## Recommendation

Work on: **🎯T1.8 Jevon imports tern's packages**
Reason: Highest effective weight (1.7) among unblocked targets. 🎯T5.1 is now achieved, and 🎯T19 is done — the library is mature. Completing the jevon migration closes 🎯T1 (the top-level "complete library" target). This is the highest-leverage next step to realize the value of all the library work done so far.

## Suggested action

Tag and push a new pigeon release if needed, then open a branch in the jevon repo to replace `internal/crypto/`, `internal/protocol/`, `internal/qr/`, and `cmd/protogen/` with imports from tern. Update the iOS app to use the `Tern` SPM package. Use `/push` to drive the PR workflow in both repos.

<!-- convergence-deps
evaluated: 2026-04-06T00:00:00Z
sha: 6ee9e4f

🎯T1:
  gap: close
  assessment: "7/8 sub-targets achieved. Only T1.8 (jevon imports) remains."
  read:
    - docs/targets.md

🎯T1.8:
  gap: not started
  assessment: "v0.9.0 released. Jevon migration pending — requires tagging tern and updating imports."
  read:
    - docs/targets.md

🎯T5:
  gap: significant
  assessment: "1/4 sub-targets achieved (T5.1). T5.2 unblocked but not started. T5.3 blocked on T10."
  read:
    - docs/targets.md

🎯T5.1:
  gap: achieved
  assessment: "ModeDatagrams implemented with full test coverage."
  read:
    - crypto/crypto.go
    - crypto/crypto_test.go

🎯T6:
  gap: not started
  assessment: "No investigation artifacts exist."
  read:
    - docs/targets.md

🎯T17:
  gap: not started
  assessment: "No Makefile deploy target. CI deploy now working after Dockerfile fix."
  read:
    - docs/targets.md
    - Dockerfile

🎯T19:
  gap: achieved
  assessment: "Hierarchical state machines landed. All generators updated, session.yaml refactored."
  read:
    - protocol/protocol.go
    - protocol/session.yaml

🎯T8:
  gap: significant
  assessment: "3/5 sub-targets achieved. T8.4 (TypeScript client) and T8.5 (LAN direct) remain."
  read:
    - docs/targets.md

standing-invariant:
  gap: green
  assessment: "Tests pass. CI green. Dockerfile fixed."
  read:
    - Dockerfile
-->
