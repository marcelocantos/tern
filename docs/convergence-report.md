# Convergence Report

Evaluated: 2026-03-28

Standing invariants: all green. CI passing (5/5 recent runs succeeded). No open PRs.

## Movement

- 🎯T13: significant -> **achieved** (runtime verification confirmed certs persist)
- 🎯T14: status updated (Playwright headless Chromium confirmed not to support WebTransport/QUIC)
- All others: (unchanged)

## Gap Report

### 🎯T1 Tern is a complete library  [weight: 1.7]
Gap: converging (7/8 sub-targets achieved)

  [x] 🎯T1.1 Crypto library — achieved
  [x] 🎯T1.2 Pairing protocol spec — achieved
  [x] 🎯T1.3 TLA+ formal model — achieved
  [x] 🎯T1.4 Protocol state machine framework — achieved
  [x] 🎯T1.5 QR helper — achieved
  [x] 🎯T1.6 Swift package — achieved
  [x] 🎯T1.7 E2E integration test — achieved
  [ ] 🎯T1.8 Jevon imports tern's packages — not started (requires tern tag + push)

### 🎯T3 Fly.io deployment via CI  [weight: 1.7]
Gap: not started
CI workflow exists (`.github/workflows/ci.yml`) but only runs `go test`. No Fly.io deploy step. Needs a deploy workflow or step that calls `flyctl deploy`.

### 🎯T5.1 Reorder-tolerant decryption  [weight: 1.7]
Gap: not started
No buffering logic exists in `Channel.Decrypt`. First step in the 🎯T5 multi-transport chain.

### 🎯T12 Channel API  [weight: 1.7]
Gap: not started (0/2 sub-targets achieved)

  [ ] 🎯T12.1 Streaming channels — not started (no OpenChannel/AcceptChannel functions exist)
  [ ] 🎯T12.2 Datagram channels — not started (blocked on 🎯T12.1)

### 🎯T16 Fly.io auto-start for UDP  [weight: 1.7]
Gap: not started
No implementation or investigation artifacts found.

### 🎯T6 Investigate STUN/NAT hole-punching  [weight: 1.5]  (status only)
Status: not started

### 🎯T17 Makefile deploy target  [weight: 1.5]  (status only)
Status: not started

### 🎯T13 Certmagic storage alignment  [weight: 2.5]
Gap: **achieved**
Verified 2026-03-28. ACME account, cert, and key persist in /data/certmagic on the Fly volume. Cert is reused on restart without re-provisioning. Moved to Achieved section.

### 🎯T7 Investigate Bluetooth as proximity oracle  [weight: 1.0]  (status only)
Status: not started

### 🎯T8 WebTransport relay  [weight: 1.0]
Gap: converging (3/5 sub-targets achieved)

  [x] 🎯T8.1 WebTransport relay server — achieved
  [x] 🎯T8.2 Non-strict Channel.Decrypt — achieved
  [x] 🎯T8.3 Go WebTransport client — achieved
  [ ] 🎯T8.4 Web/TypeScript WebTransport client — not started
  [ ] 🎯T8.5 LAN direct WebTransport — not started (blocked on 🎯T8.4)

### 🎯T14 Browser WebTransport E2E  [weight: 1.0]  (status only)
Status: blocked on Playwright headless Chromium QUIC support. LE rate limit block has expired. Server config verified correct. Options: headed Chrome, Selenium, or manual verification.

### 🎯T15 Gomobile bindings  [weight: 1.0]  (status only)
Status: not started

### 🎯T5 Multi-transport with LAN upgrade  [weight: 0.8]
Gap: not started (0/4 sub-targets achieved)
All sub-targets blocked or not started. 🎯T5.3 blocked on 🎯T10 (TLA+ model).

### 🎯T10 TLA+ model for cutover protocol  [weight: 0.6]
Gap: not started (status only)
Low weight (cost exceeds value ratio). Consider whether this is needed before 🎯T5.3 implementation or can be done in parallel.

## Recommendation

Work on: **🎯T3 Fly.io deployment via CI**
Reason: Tied at 1.7 effective weight with T1.8, T5.1, T12.1, and T16. T3 is the highest-leverage choice because it eliminates manual deploy friction for every subsequent change, directly supports T16 (auto-start investigation benefits from automated deploys), and is a prerequisite mindset for T17 (Makefile deploy target). Low risk, well-understood scope (add `flyctl deploy` step to existing CI workflow).

## Suggested action

Add a deploy job to `.github/workflows/ci.yml` that runs after tests pass on the `master` branch. Use `superfly/flyctl-setup-action` and `flyctl deploy --remote-only`. Store the Fly API token as a GitHub secret (`FLY_API_TOKEN`). Gate the deploy step on `github.ref == 'refs/heads/master'` so PRs only run tests.

<!-- convergence-deps
evaluated: 2026-03-28T12:00:00Z
sha: 68118b3

🎯T13:
  gap: achieved
  assessment: "Verified done. Moved to Achieved section."
  read:
    - docs/targets.md

🎯T1:
  gap: close
  assessment: "7/8 sub-targets achieved. Only T1.8 (jevon imports) remains — requires tagging."
  read:
    - docs/targets.md

🎯T3:
  gap: not started
  assessment: "CI runs tests only. No fly deploy step."
  read:
    - .github/workflows/ci.yml

🎯T12:
  gap: not started
  assessment: "No OpenChannel/AcceptChannel functions exist."
  read:
    - conn.go
    - tern.go

🎯T5.1:
  gap: not started
  assessment: "No buffering logic in Channel.Decrypt."
  read:
    - crypto/crypto.go

🎯T16:
  gap: not started
  assessment: "No implementation or investigation artifacts found."
  read:
    - fly.toml

🎯T14:
  gap: significant
  assessment: "Playwright headless Chromium doesn't support WebTransport. Need alternative approach."
  read:
    - docs/targets.md
    - webtransport.go
    - cmd/tern/main.go
-->
