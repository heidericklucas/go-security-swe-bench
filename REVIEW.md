# Code Review: Candidate Patches

This benchmark grades a candidate patch with a binary reward: does it flip the
instance's exploit test red→green while keeping every functional contract green?
That reward is **necessary but not sufficient**. A patch can earn **reward 1.0 and
still be wrong** — it closes the single exploit the grader ships while leaving the
vulnerability *class* open. Catching those is the job of human review, and it is
the skill this document demonstrates.

Every patch below is real and reproducible:

```bash
python3 harness/run_eval.py --instance <id> --patch docs/review-samples/<patch>
```

---

## 1. Passes the grader, fails review — **REJECT**

### `pathtraversal-hasprefix.patch` → reward **1.0**, still exploitable

The patch confines reads with a string comparison:

```go
p := filepath.Clean(filepath.Join(base, name))
if !strings.HasPrefix(p, base) { return nil, os.ErrPermission }
```

It blocks the shipped exploit — `../app_secret.txt` resolves outside `base`, so
`HasPrefix` fails — and scores **reward 1.0**. But `HasPrefix` compares *strings*,
not *path boundaries*:

- **Sibling-prefix escape.** With `base = /data/att`, the path `/data/att-evil/secret`
  satisfies `HasPrefix(…, "/data/att")` and is served — a different directory that
  merely shares the name prefix.
- **Symlink escape.** The check runs on the lexical path; a symlink *inside* `base`
  pointing outside is accepted by the string check and then followed by `os.Open`.

**Missing coverage to add:** a `FAIL_TO_PASS` requesting a sibling-prefix path
(`../<attachments>-evil/secret`) and a symlink-escape case. The golden
(`os.OpenRoot`) already defeats both — confinement is enforced by the kernel at the
file-descriptor level, not by comparing strings.

### `xss-blocklist.patch` → reward **1.0**, trivially bypassable

The patch strips one tag before rendering with `text/template`:

```go
n.Body = strings.ReplaceAll(n.Body, "<script>", "")
n.Body = strings.ReplaceAll(n.Body, "</script>", "")
```

The shipped exploit injects exactly `<script>alert(1)</script>`, which the strip
removes — so it scores **reward 1.0**. But a blocklist on one tag is not output
encoding. All of these survive and execute:

- `<img src=x onerror="alert(1)">` — no `<script>` tag at all.
- `<ScRiPt>alert(1)</ScRiPt>` — a case the exact-match replace never touches.
- `<scr<script>ipt>alert(1)` — removing the inner `<script>` rejoins `<scr` + `ipt>`
  into a fresh `<script>`, because `ReplaceAll` is single-pass.

**Missing coverage to add:** `FAIL_TO_PASS` cases for an event-handler payload and a
mixed-case payload. The golden (`html/template`) needs none of them — it escapes by
output *context*, so every payload above renders as inert text.

---

## 2. Golden solutions — **ACCEPT**

Each golden closes the vulnerability *class*, not just the shipped exploit, and
preserves every functional contract (`make verify-solution` → reward 1.0 on all six).

| Instance | Fix | Why it is complete |
|---|---|---|
| `ssrf-import` | Per-hop IP classification in the custom `DialContext` | Blocks the resolved IP on **every** hop → TOCTOU/rebind-immune (a pre-fetch `LookupIP` is not). The denylist (loopback/private/link-local/…) is not range-exhaustive; a strict allowlist would be the airtight form. |
| `jwt-forged-token` | `jwt.Parse` + `WithValidMethods(RS256)` + `WithExpirationRequired` | Rejects the wrong-key forgery *by signature*; no claim/alg heuristic can substitute |
| `sqli-note-search` | Parameterized `?` placeholders | User input is bound as a value; it can never re-enter SQL grammar |
| `idor-note-read` | Owner check before returning the note | Unconditional `403` on ownership mismatch |
| `xss-note-export` | `text/template` → `html/template` | Context-aware auto-escaping; no blocklist to bypass |
| `pathtraversal-attachment` | `os.OpenRoot` confinement | Kernel-enforced boundary; defeats sibling-prefix and symlink escapes |

---

## 3. Cross-cutting findings

**Performance — note search is a full scan.** `SearchNotes` filters with
`title LIKE ?` against `notes` with no index on `(owner_id, title)`; each search is
O(rows). Harmless at seed scale, but a real deployment wants a composite index (or
FTS) — noted so a *correct* security fix is not mistaken for a *performant* one.

**Maintainability — the ownership check is duplicated and inconsistent.** The same
`claimsFrom(r).UserID != n.OwnerID` guard is repeated across `getNote`, `deleteNote`,
and `exportNote` — and it diverges: `getNote`/`deleteNote` return `403`, while
`exportNote` folds the mismatch into its not-found branch and returns `404`. Extract a
`requireOwner(r, n) error` helper so the rule *and its status code* live in one place
and a future endpoint can't forget it or answer differently.

---

## 4. Takeaway

The automated reward proves a patch closes the **specific** exploit shipped with the
instance. It cannot prove the patch closes the **class** — the two REJECTs above both
earn reward 1.0 and both ship the vulnerability. That gap is the point: each
instance's `FAIL_TO_PASS` set should keep growing toward the class, and a human
review gate belongs on top of the reward, not under it.
