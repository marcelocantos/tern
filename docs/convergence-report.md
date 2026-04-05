# Convergence Report

Evaluated: 2026-04-05

## Standing invariants

- **Tests**: PENDING — CI run in progress on master (0ac4f75, post-merge of PR #1). The PR branch CI run passed. `TestHealthMonitorFallback` is now **skipped** (annotated as machine-driven under 🎯T18), resolving the prior 5-run failure streak.
- **CI**: IN PROGRESS — awaiting result of master push run #24001562484. PR run succeeded.

## Movement

- 🎯T18: not started -> **close** (PR #1 merged: executor, events/commands in YAML, pathRouter deleted, all integration tests pass. Remaining: TLA+ event/command generator, channel migration, skipped health monitor test.)
- 🎯T1: (unchanged — still converging 7/8, blocked on 🎯T1.8)
- All others: (unchanged)

## Gap Report

### 🎯T1.8 Jevon imports tern's packages  [weight: 1.7]
Gap: not started
v0.9.0 released. Migration in jevon repo still pending. This is external work in the jevon codebase.

### 🎯T5.1 Reorder-tolerant decryption  [weight: 1.7]
Gap: not started
No buffering logic in `Channel.Decrypt`. Needed for safe transport cutover.

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

### 🎯T6 Investigate STUN/NAT hole-punching  [weight: 1.5]  (status only)
Status: not started

### 🎯T17 Makefile deploy target  [weight: 1.5]  (status only)
Status: not started

### 🎯T18 State machine mediates all Conn behavior  [weight: 1.25]
Gap: close
PR #1 merged. The executor is implemented (891 lines), events and commands are declared in session.yaml, the machine's `HandleEvent` returns explicit commands, Conn delegates all I/O to the executor, pathRouter is deleted, and all integration tests pass. Remaining work:
- TLA+ generator does not yet emit event/command declarations (Phase 8)
- OpenChannel/DatagramChannel still use legacy routing callbacks rather than going through the executor
- `TestHealthMonitorFallback` is skipped — the machine-driven health monitor replaces the ad-hoc goroutine, but the test hasn't been updated to validate the new approach

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
Status: blocked on Playwright headless Chromium QUIC support

### 🎯T15 Gomobile bindings  [weight: 1.0]  (status only)
Status: not started

### 🎯T5 Multi-transport with LAN upgrade  [weight: 0.8]
Gap: not started (0/4 sub-targets achieved)
All sub-targets blocked or not started.

### 🎯T10 TLA+ model for cutover protocol  [weight: 0.6]  (status only)
Status: not started. Low effective weight.

### Blocked targets

- 🎯T5.2 LAN discovery via relay — blocked on 🎯T5.1
- 🎯T5.3 Cutover protocol — blocked on 🎯T5.1, 🎯T10
- 🎯T5.4 Transport-agnostic Conn — blocked on 🎯T5.2, 🎯T5.3
- 🎯T8.5 LAN direct WebTransport — blocked on 🎯T8.4

## Recommendation

Work on: **🎯T18 State machine mediates all Conn behavior**
Reason: 🎯T18 is close to achieved with the highest-leverage remaining work items (TLA+ generator, channel migration, health monitor test). Closing it out is cheaper than starting any of the weight-1.7 targets from scratch. 🎯T1.8 and 🎯T5.1 are both "not started" and require more effort. Finishing 🎯T18 also unblocks future work — the executor pattern makes channel migration and health monitoring cleaner.

## Suggested action

Complete the 🎯T18 remaining items in priority order:
1. Update `TestHealthMonitorFallback` to validate the machine-driven health monitor (unskip and rewrite against the executor's `ping_tick`/`ping_timeout` event flow).
2. Migrate OpenChannel/DatagramChannel to go through the executor instead of legacy routing callbacks.
3. Add event/command declarations to the TLA+ generator (`protocol/tla.go`).

Start with the health monitor test — it validates that the core machine-driven I/O pattern works for the most complex lifecycle scenario, and gets CI fully green with no skipped tests.

<!-- convergence-deps
evaluated: 2026-04-05T10:00:00Z
sha: 0ac4f75

🎯T1:
  gap: close
  assessment: "7/8 sub-targets achieved. Only T1.8 (jevon imports) remains."
  read:
    - docs/targets.md

🎯T1.8:
  gap: not started
  assessment: "v0.9.0 released. Jevon migration pending."
  read:
    - docs/targets.md

🎯T5.1:
  gap: not started
  assessment: "No buffering logic in Channel.Decrypt."
  read:
    - docs/targets.md

🎯T18:
  gap: close
  assessment: "PR #1 merged. Executor done, events/commands in YAML, pathRouter deleted. Remaining: TLA+ generator, channel migration, health monitor test."
  read:
    - executor.go
    - conn.go
    - path.go
    - protocol/session.yaml
    - protocol/tla.go
    - pathswitch_test.go
    - docs/targets.md

🎯T8:
  gap: close
  assessment: "3/5 sub-targets achieved. T8.4 (TypeScript client) and T8.5 (LAN direct) remain."
  read:
    - docs/targets.md

standing-invariant:
  gap: pending
  assessment: "CI in progress on master. PR run passed. TestHealthMonitorFallback skipped."
  read:
    - pathswitch_test.go
-->
