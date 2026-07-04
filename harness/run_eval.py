#!/usr/bin/env python3
"""Grade a candidate patch against a task instance and emit a scalar reward.

Hermetic: runs `go test` inside a pinned golang container with no network and a
vendored module cache. Stdlib only.
"""
import argparse
import glob
import json
import os
import subprocess
import sys
import tempfile

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
GO_IMAGE = "golang:1.24-bookworm@sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac"
# Candidate patches may ONLY modify app/. Everything else is held out and reset
# from canonical after applying.
PROTECTED_PREFIXES = ("tests/", "tasks/", "harness/", "solution/", ".github/")
PROTECTED_EXACT = ("go.mod", "go.sum", "Makefile")


def sh(cmd, **kw):
    return subprocess.run(cmd, capture_output=True, text=True, **kw)


def patched_paths(patch_text):
    # Collect target paths from BOTH sides so deletions (+++ /dev/null) and
    # renames are still scope-checked, not just in-place edits.
    paths = []
    for line in patch_text.splitlines():
        if line.startswith("+++ b/"):
            paths.append(line[len("+++ b/"):].strip())
        elif line.startswith("--- a/"):
            paths.append(line[len("--- a/"):].strip())
    return paths


def in_scope(paths):
    for p in paths:
        if p in PROTECTED_EXACT or p.startswith(PROTECTED_PREFIXES):
            return False, p
        if not p.startswith("app/"):
            return False, p
    return True, None


def build_workspace(patch_file):
    """Extract canonical HEAD, apply the (scoped) patch, reset held-out files."""
    tmp = tempfile.mkdtemp(prefix="eval-")
    # Extract canonical HEAD via binary archive (avoids text mangling).
    arch = subprocess.run(["git", "-C", REPO, "archive", "HEAD"], capture_output=True)
    subprocess.run(["tar", "-x", "-C", tmp], input=arch.stdout)

    scope_ok = True
    if patch_file and os.path.getsize(patch_file) > 0:
        text = open(patch_file, encoding="utf-8", errors="replace").read()
        ok, bad = in_scope(patched_paths(text))
        if not ok:
            return tmp, False, f"patch touches protected path: {bad}"
        ap = sh(["git", "apply", "-p1", os.path.abspath(patch_file)], cwd=tmp)
        if ap.returncode != 0:
            return tmp, False, f"git apply failed: {ap.stderr.strip()}"
    # belt-and-suspenders: reset every held-out path from canonical
    held = ["tests", "tasks", "harness", "go.mod", "go.sum", "Makefile", ".github"]
    reset = subprocess.run(["git", "-C", REPO, "archive", "HEAD", *[h for h in held if path_exists_in_head(h)]], capture_output=True)
    subprocess.run(["tar", "-x", "-C", tmp], input=reset.stdout)
    return tmp, scope_ok, None


def path_exists_in_head(p):
    return sh(["git", "-C", REPO, "cat-file", "-e", f"HEAD:{p}"]).returncode == 0 or \
           sh(["git", "-C", REPO, "ls-tree", "HEAD", p]).stdout.strip() != ""


def run_tests(workspace, packages, names):
    run_re = "^(" + "|".join(names) + ")$"
    cmd = [
        "docker", "run", "--rm", "--network", "none",
        "-v", f"{workspace}:/src", "-w", "/src",
        "-v", "gosec-gocache:/gocache",  # shared build cache across grades (speed only)
        "-e", "GOCACHE=/gocache",
        "-e", "GOFLAGS=-mod=vendor", "-e", "GOPROXY=off", "-e", "CGO_ENABLED=0",
        GO_IMAGE, "go", "test", "-json", "-count=1", "-run", run_re, *packages,
    ]
    res = sh(cmd)
    observed = {}
    for line in res.stdout.splitlines():
        try:
            ev = json.loads(line)
        except json.JSONDecodeError:
            continue
        if ev.get("Test") and ev.get("Action") in ("pass", "fail", "skip"):
            observed[ev["Test"]] = ev["Action"]
    return observed, res


def grade(instance, patch_file):
    m = json.load(open(os.path.join(REPO, "tasks", f"{instance}.json")))
    ws, ok, err = build_workspace(patch_file)
    if not ok:
        return {"instance": instance, "reward": 0.0, "error": err}
    ftp, ptp = m["fail_to_pass"], m["pass_to_pass"]
    observed, res = run_tests(ws, m["test_packages"], ftp + ptp)
    fail_detail = {n: observed.get(n, "MISSING") for n in ftp + ptp}
    reward = 1.0 if all(observed.get(n) == "pass" for n in ftp + ptp) else 0.0
    out = {"instance": instance, "reward": reward, "fail_to_pass": {n: fail_detail[n] for n in ftp},
           "pass_to_pass": {n: fail_detail[n] for n in ptp}}
    if reward == 0 and res.returncode != 0 and not observed:
        out["error"] = "build/test error:\n" + res.stderr[-2000:]
    return out


def all_instances():
    return sorted(os.path.splitext(os.path.basename(p))[0] for p in glob.glob(os.path.join(REPO, "tasks", "*.json")))


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--all", action="store_true")
    ap.add_argument("--base", action="store_true")
    ap.add_argument("--solution", action="store_true")
    ap.add_argument("--attacks", action="store_true")
    ap.add_argument("--instance")
    ap.add_argument("--patch")
    a = ap.parse_args()

    if a.instance and a.patch is not None:
        print(json.dumps(grade(a.instance, a.patch), indent=2))
        return

    failures = 0
    if a.attacks:
        man = json.load(open(os.path.join(REPO, "tests", "adversarial", "manifest.json")))
        for entry in man["patches"]:
            r = grade(entry["instance"], os.path.join(REPO, entry["patch"]))
            bad = r["reward"] != 0
            failures += bad
            print(f"[{'FAIL' if bad else 'ok'}] attack {entry['patch']} -> reward {r['reward']} ({entry['note']})")
        sys.exit(1 if failures else 0)

    expect = 1 if a.solution else 0
    for inst in all_instances():
        patch = os.path.join(REPO, json.load(open(os.path.join(REPO, "tasks", f"{inst}.json")))["golden"]) if a.solution else None
        r = grade(inst, patch)
        bad = r["reward"] != expect
        failures += bad
        print(f"[{'FAIL' if bad else 'ok'}] {inst} -> reward {r['reward']} (expected {expect})")
        if bad and "error" in r:
            print("    " + r["error"].replace("\n", "\n    "))
    sys.exit(1 if failures else 0)


if __name__ == "__main__":
    main()
