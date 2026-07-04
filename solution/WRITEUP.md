# WRITEUP: Exploit Narratives and Golden Fixes

Each task instance is a single-defect vulnerability repair challenge. This document explains the vulnerability, attack strategy, and why the golden solution works.

---

## Instance 1: SSRF via Redirect Following (CWE-918, A10:2021)

**ID**: `ssrf-import`

### Attack Narrative

The base fetcher performs no IP validation on any hop, so a request to an internal address is dialed directly, and a request to an allowed external host that **redirects** to an internal one is followed there too (e.g. cloud metadata at `169.254.169.254`). The instance ships two exploits — a direct-internal fetch and a redirect-to-internal fetch — to prove enforcement on both.

### Golden Reasoning

The golden patch inserts an IP-classification check (`blockedIP`) **inside the HTTP dialer**, ensuring every TCP dial (initial request + redirect hops) invokes the check. This prevents DNS rebinding attacks and TOCTOU vulnerabilities.

---

## Instance 2: JWT Forged Token (CWE-347, A07:2021)

**ID**: `jwt-forged-token`

### Attack Narrative

The base code calls `jwt.NewParser().ParseUnverified(token, &claims)`, skipping signature verification entirely. An attacker forges a JWT with `{"uid": 999, "admin": true}` and gains unauthorized access.

### Golden Reasoning

The golden patch calls `jwt.Parse()` with proper signature verification, RSA public key validation, RS256 method enforcement, and expiration checking. Only legitimate tokens are accepted.

---

## Instance 3: SQL Injection in Search (CWE-89, A03:2021)

**ID**: `sqli-note-search`

### Attack Narrative

The base code builds the query with `fmt.Sprintf`, so `q` lands directly inside the SQL. A `UNION SELECT` payload that closes the `LIKE` string literal and comments out the trailing clause leaks another owner's rows past the `owner_id` filter — this is exactly the payload the shipped exploit uses.

### Golden Reasoning

The golden patch uses parameterized queries with placeholders (`?`), binding user input as data instead of SQL code. Pattern construction moves to application logic: `"%"+q+"%"`.

---

## Instance 4: Broken Object-Level Access Control (CWE-639, A01:2021)

**ID**: `idor-note-read`

### Attack Narrative

The endpoint returns notes by ID without ownership verification. An attacker guesses another user's note ID and reads it.

### Golden Reasoning

The golden patch adds an immediate ownership check: `if claimsFrom(r).UserID != n.OwnerID { return 403 }`. Every resource access is gated by authorization.

---

## Instance 5: Stored XSS — Missing Output Encoding (CWE-79, A03:2021)

**ID**: `xss-note-export`

### Attack Narrative

The export endpoint renders a note's stored content as HTML without escaping. A note whose body contains `<img src=x onerror="alert('XSS')">` executes that script when its owner opens the export view — the payload is persisted, so it fires on every render. (Export is owner-gated, so this is a stored self-XSS, not a cross-user one.)

### Golden Reasoning

The golden patch switches from `text/template` to `html/template`, enabling auto-escaping of HTML special characters. User content is safely neutralized.

---

## Instance 6: Path Traversal in File Download (CWE-22, A01:2021)

**ID**: `pathtraversal-attachment`

### Attack Narrative

The endpoint uses `filepath.Join(baseDir, name)` which cleans but does not confine the path. An attacker supplies `../../etc/passwd` to read system files.

### Golden Reasoning

The golden patch uses `os.OpenRoot(baseDir)` to create a confined file descriptor rooted at the base directory. All file operations are kernel-enforced within bounds.

---

## Golden Patches Summary

| Instance | Fix | Principle |
|---|---|---|
| SSRF | Dialer-level IP blocking | Validate on every hop |
| JWT | Full signature verification + RS256 | Parse with verification |
| SQLi | Parameterized queries | Bind as data, not code |
| IDOR | Ownership check | Verify authorization |
| XSS | html/template (auto-escape) | Auto-escape by default |
| Path Traversal | os.OpenRoot confinement | OS-level boundaries |
