# Task Instance Manifest Schema

Each task instance is defined by a JSON file at `tasks/<id>.json`. The grader (`harness/run_eval.py`) reads these manifests to determine which tests must pass.

## Fields

- **`id`** (string, required): Unique identifier for the task instance (e.g., `ssrf-import`).
- **`title`** (string, required): Human-readable task title.
- **`cwe`** (string, required): CWE identifier(s) associated with the bug.
- **`owasp`** (string, required): OWASP Top 10 category (e.g., `A10:2021`).
- **`problem_statement_file`** (string, required): Path to the model-visible problem statement (e.g., `tasks/ssrf-import.md`). This file is shown to the model; it leads with the observable symptom and names no CWE/OWASP code or fix.
- **`test_packages`** (array of strings, required): Go test package paths to run (e.g., `["./tests/exploit/...", "./tests/functional/..."]`).
- **`fail_to_pass`** (array of test names, required): Test IDs that currently fail in the base codebase. A successful fix must make all of these pass.
- **`pass_to_pass`** (array of test names, required): Test IDs that currently pass in the base codebase. A successful fix must not break any of these.
- **`golden`** (string, required): Path to the canonical solution patch (relative to repo root, e.g., `solution/ssrf-import.patch`). Used for verification.
- **`reward_type`** (string, required): Grading mode (currently only `binary` is supported: `0.0` if any test fails, `1.0` if all tests pass).

## Grading Contract

The grader runs `go test -json` inside a hermetic Docker container. It:

1. Extracts the canonical HEAD repository state.
2. Applies the candidate patch (scoped to `app/` only; other paths reset from canonical).
3. Runs all tests in `fail_to_pass` and `pass_to_pass`.
4. Emits `reward = 1.0` iff ALL tests are observed to pass; otherwise `reward = 0.0`.

Failures due to build errors, missing tests, or test timeouts all yield reward 0.

## Adding a Task

To create a new task:

1. Write the model-visible problem statement in `tasks/<id>.md` (symptom-first, no CWE/OWASP/fix hints).
2. Create the manifest `tasks/<id>.json` with the schema above. Identify the currently-failing tests (`fail_to_pass`) and the tests that should remain passing (`pass_to_pass`).
3. Create a reference solution patch at `solution/<id>.patch` (generated via `git diff HEAD`).
4. Test the grader: `python3 harness/run_eval.py --instance <id> --patch /dev/null` should emit `reward 0.0`; `--patch solution/<id>.patch` should emit `reward 1.0`.
