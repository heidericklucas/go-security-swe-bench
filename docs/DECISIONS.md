# DECISIONS.md: Architecture Decision Records

---

## ADR-1: `make test` Is Green on the Base App

**Status**: Accepted

**Decision**: `make test` runs `./app/...` + `./tests/functional/...` and **passes on the vulnerable base**. The held-out exploit tests (`tests/exploit/`) are excluded from `make test` and run only through the grader.

**Rationale**: `make test` is a fast, honest "the app works" signal for contributors — it must not be red on a correct base. The security defects are demonstrated by the grader flipping each instance from reward 0 (base) to reward 1 (golden), not by a failing `make test`. Unit tests under `app/` (e.g. `safepath_test.go`) assert legitimate behavior that holds in both the vulnerable and fixed code, so they stay green throughout.

---

## ADR-2: SSRF Check in Dialer, Not Pre-Fetch

**Status**: Accepted

**Decision**: IP blocking lives inside `dialContext`, not pre-fetch validation.

**Rationale**: `http.Transport` re-invokes dialer per redirect hop automatically. Dialer-level checks prevent TOCTOU, DNS rebinding, and redirect bypass.

---

## ADR-3: XSS Prevention via html/Template

**Status**: Accepted

**Decision**: Switch from `text/template` to `html/template`.

**Rationale**: Auto-escaping is safer than custom logic. No additional code required; stdlib is maintained and proven.

---

## ADR-4: JWT Verification with Key Function + RS256 Enforcement

**Status**: Accepted

**Decision**: Use `jwt.Parse(token, keyFunc, ...)` with `WithValidMethods([]string{"RS256"})`.

**Rationale**: The base bug is that `ParseUnverified` skips signature checking entirely, so any forged token is honored. The golden verifies the RSA signature; pinning the method with `WithValidMethods` and requiring `exp` also close the adjacent gaps (unexpected algorithms, missing expiry). Only tokens this service actually signed, and that are unexpired, are accepted.

---

## ADR-5: SQL Injection via Parameterized Queries

**Status**: Accepted

**Decision**: Use `database/sql` placeholders (`?`) for all user input.

**Rationale**: Placeholders separate SQL structure from data at driver level. User input is always treated as data, never SQL code.

---

## ADR-6: IDOR Prevention via Explicit Ownership Check

**Status**: Accepted

**Decision**: Add `if claimsFrom(r).UserID != n.OwnerID { return 403 }` immediately after retrieval.

**Rationale**: Explicit guard clauses are clearer than implicit filtering. Ownership is verified unconditionally.

---

## ADR-7: Path Traversal via os.OpenRoot Confinement

**Status**: Accepted

**Decision**: Use `os.OpenRoot(baseDir)` to create a confined file descriptor.

**Rationale**: Kernel-enforced boundaries are more robust than application logic. Impossible to escape via `../..` patterns.

---

## ADR-8: Binary Reward Grading (0 or 1.0)

**Status**: Accepted

**Decision**: Reward = 1.0 iff all tests pass; else 0.0.

**Rationale**: Security is binary; partial fixes should not earn credit. Adversarial patches (reject-all, etc.) correctly score 0.

---

## ADR-9: Model Visibility: Problem Statements Only

**Status**: Accepted

**Decision**: Models see problem statement (symptoms) + read-only app/ source. No exploit code, golden patches, or grader internals visible.

**Rationale**: Models should solve based on understanding, not pattern matching. Grader internals are not part of the evaluation scope.

---

## ADR-10: Hermetic Docker Grading

**Status**: Accepted

**Decision**: All grading via `harness/run_eval.py` in Docker with vendored dependencies.

**Rationale**: Ensures reproducibility, determinism, and consistency across environments. Build errors/timeouts yield reward 0.

---

## ADR-11: Password Storage Is Out of Scope

**Status**: Accepted

**Decision**: The demo app stores and compares passwords in plaintext and does not hash them. This is a deliberate boundary, not one of the six graded defects.

**Rationale**: Each instance targets one clean, gradable vulnerability class; authentication is reduced to the minimum needed to exercise the notes API (a signed JWT issued after a trivial credential check). Hashing would add a second, ungraded concern to the auth path and blur which defect an instance tests. A production build would hash with a memory-hard KDF and never expose the credential.
