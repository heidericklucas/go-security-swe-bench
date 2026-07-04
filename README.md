# go-security-swe-bench

A SWE-bench-style RL environment for evaluating AI models on realistic security vulnerability repair tasks.

## What

This benchmark contains 6 single-defect Go application instances, each posing a realistic security vulnerability (SSRF, JWT forgery, SQLi, IDOR, XSS, path traversal). Models are given the problem statement and vulnerable source code, then asked to fix it. Performance is evaluated binary: fix passes all tests (reward 1.0) or it doesn't (reward 0.0).

## Why

Security vulnerabilities are a leading threat vector in production software, yet few benchmarks evaluate AI safety-relevant repair workflows. This benchmark tests:
- Ability to identify and fix real exploit patterns
- Resistance to "fixes" that break legitimate functionality
- Robustness against adversarial patches that appear to work but miss critical cases

## How

### Build

```bash
make build
```

Offline compile check of every package (`go build -mod=vendor ./...`). The grader itself builds nothing from this — it runs `go test` from source inside the pinned container.

### Run Tests (Offline)

```bash
make test
```

Runs the unit and functional tests on the base app. These **pass** on the base — the app's legitimate behavior is intact. The security defects are proven separately by the grader's held-out exploit tests (via `make verify`), never by `make test`.

### Grade the Base Application (Expect Reward 0)

```bash
make verify
```

Runs the harness in Docker, grading the base app against all 6 instances. All should fail (no fixes applied).

### Grade with Golden Patches (Expect Reward 1.0)

```bash
make verify-solution
```

Applies each canonical solution patch and grades. All instances should pass.

### Grade Adversarial Patches (Expect Reward 0)

```bash
make verify-attacks
```

Applies intentionally broken "fixes" to verify the test suite catches insufficient repairs. All should fail.

### Coverage

```bash
make cover
```

Reports per-function coverage of the `app/` code. This runs in the repo's own CI only — it is **never** part of the reward function.

## Instances

Each instance has:
- A **problem statement** (visible to the model): `tasks/<id>.md`
- A **manifest** with test contracts: `tasks/<id>.json`
- A **golden patch**: `solution/<id>.patch`

| ID | Vulnerability | CWE | OWASP | Problem Statement |
|---|---|---|---|---|
| `ssrf-import` | SSRF via redirect following | CWE-918 | A10:2021 | [tasks/ssrf-import.md](tasks/ssrf-import.md) |
| `jwt-forged-token` | JWT forgery (unverified signature) | CWE-347 | A07:2021 | [tasks/jwt-forged-token.md](tasks/jwt-forged-token.md) |
| `sqli-note-search` | SQL injection in note search | CWE-89 | A03:2021 | [tasks/sqli-note-search.md](tasks/sqli-note-search.md) |
| `idor-note-read` | Broken object-level access control | CWE-639 | A01:2021 | [tasks/idor-note-read.md](tasks/idor-note-read.md) |
| `xss-note-export` | Stored XSS (missing output encoding) | CWE-79 | A03:2021 | [tasks/xss-note-export.md](tasks/xss-note-export.md) |
| `pathtraversal-attachment` | Path traversal in file download | CWE-22 | A01:2021 | [tasks/pathtraversal-attachment.md](tasks/pathtraversal-attachment.md) |

## Model Evaluation

The harness presents each model with:
1. A problem statement (symptom-first, no fix hints)
2. Read-only access to `app/` source
3. An expected list of failing tests
4. The model must patch `app/` to make all tests pass

The model **cannot see**:
- Exploit code in `tests/exploit/`
- Golden patches in `solution/`
- This README or any grader internals

Fixes are evaluated deterministically in Docker:
- Binary reward (0.0 or 1.0)
- All `fail_to_pass` tests must pass
- All `pass_to_pass` tests must not break
- Build errors or timeouts yield reward 0

## Documentation

- [REVIEW.md](REVIEW.md) — code review of candidate patches, including patches that **pass the automated grader (reward 1.0) yet are still wrong**, plus performance and maintainability findings
- [DECISIONS.md](docs/DECISIONS.md) — architecture decision records
- [solution/WRITEUP.md](solution/WRITEUP.md) — per-instance exploit narratives and golden reasoning (grader-only; not shown to the model)

## Repository Structure

```
.
├── app/                    # Application source (Go packages; the model patches this)
├── tests/
│   ├── exploit/            # Exploit proofs — FAIL_TO_PASS (grader-only)
│   ├── functional/         # Positive contracts + happy path — PASS_TO_PASS
│   ├── harness/            # Shared in-process test app (held-out)
│   └── adversarial/        # Reward-hacking patches proven to score 0
├── tasks/                  # Instance manifests (*.json) + model-visible statements (*.md)
├── solution/               # Golden patches + WRITEUP (grader-only)
├── harness/                # Python grader (run_eval.py) + schema
├── docs/                   # DECISIONS.md and review-samples/
├── REVIEW.md               # Code review of candidate patches
├── Makefile
└── go.mod, go.sum, vendor/ # Vendored deps (hermetic, offline build)
```

## License

MIT — see [LICENSE](LICENSE).
