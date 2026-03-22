# Audit Log

Chronological record of audits, releases, documentation passes, and other
maintenance activities. Append-only — newest entries at the bottom.

## 2026-03-22 — /open-source tern v0.1.0

- **Commit**: `6782a9c`
- **Outcome**: Open-sourced tern. Migrated all library code from jevon (crypto, protocol framework, QR helper, protogen tool, Swift package). Audit: 19 findings across security, correctness, docs, and CI — all critical/high addressed. Docs: README, CLAUDE.md, agents-guide.md, inline doc comments, pairing ceremony SVG diagram. v0.1.0 release pending push.
- **Deferred**:
  - Protocol actor names still use private names ("jevond", "ios") in pairing.yaml and generated files — display label fixed in puml/SVG only
  - Protocol framework `Example_test.go` (Priority 4)
  - Swift confirmation code documentation (Priority 4 — depends on Swift getting DeriveConfirmationCode)
  - CI workflow not yet set up (needed before v0.1.0 release)
