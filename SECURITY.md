# Security assessment

Assessed on **2026-11-04** against commit
`fabdc8c51b097c0fdbf77c0594e030fd256fc6db` using OWASP Top 10:2025.
This is a point-in-time assessment, not a certification or a guarantee that no
other vulnerabilities exist. The initial assessment refers to the commit above;
follow-up mitigation status is recorded separately below.

## Scope and outcome

Reviewed the Go HTTP handlers, authentication, authorization, persistence,
image downloading/processing, embedded HTML/JavaScript, runtime configuration,
and CI/release controls. Tested locally using disposable data and the existing
HTTPS browser harness, which exercises the production handler. No public
deployment, production household, real S3 account, or external target was tested.

**No critical or high-severity issue was established in this scope.** One
availability weakness was reproduced. Session lifetime and first-run ownership
need attention; transport and browser hardening depend on deployment choices.
Severity below is qualitative and conditional on the stated prerequisites.

| ID | Initial severity | Finding | Evidence | Current status |
| --- | --- | --- | --- | --- |
| SEC-01 | Medium | Unauthenticated partial POST blocks application requests | Local HTTP reproduction and code | Mitigated locally; review pending |
| SEC-02 | Medium | Admin sessions have no server-side time expiry | Static review | Open |
| SEC-03 | Medium, conditional | First reachable visitor can initialize an unconfigured instance | Public setup page observed; setup flow reviewed and regression-tested | Open |
| SEC-04 | Medium, conditional | Remote S3 endpoint may use unencrypted HTTP | Static review | Open |
| SEC-05 | Low | CSP provides framing protection only | Response headers and static review | Open |

## Findings and recommended actions

### SEC-01: Request parsing holds the global application mutex

`internal/app/admin.go:149` acquires `a.mu` before body parsing at line 172;
authentication and CSRF rejection happen afterward. A client can send a POST
with a declared body length and leave the body incomplete, holding the mutex
while other household requests wait. `cmd/wishmas/main.go` bounds reads to
15 seconds, but repeated requests can sustain contention. The actual impact
behind a buffering/rate-limiting reverse proxy was not measured.

Reproduction used a disposable production binary on loopback port 18081: send
`POST /login` with `Content-Type: application/x-www-form-urlencoded`,
`Content-Length: 100`, and only `x=` as body. After 300 ms, a concurrent `GET /`
timed out at two seconds. `GET /healthz` returned 200; after closing the partial
POST connection, `GET /` returned 200 again. Only one incomplete request was
used; no load or sustained denial-of-service test was performed.

**Recommendation:** bound and parse request bodies before acquiring household
locks; minimize lock scope around state changes. Add a regression test proving
an incomplete unauthenticated body does not block an independent read. Apply
proxy body deadlines and rate limits as defense in depth. Liveness alone must
not be interpreted as functional readiness.

**Follow-up — 2026-11-05:** the uncommitted OPS-001-AC7 change moves bounded
POST parsing ahead of the application lock, while keeping error rendering after
household loading to preserve locale and storage-error precedence. The existing
16 KiB limit, server read deadline, authentication/CSRF checks and serialized
state changes remain. A failing-first regression reproduced blocking before and
after setup; both cases now pass, alongside malformed/oversized request checks.
The original raw HTTP scenario was repeated against a fresh production binary:
an incomplete POST remained pending while the setup page and health probe each
returned 200. Go race checks, lint, static builds and all 22 browser tests pass.
This mitigates the reproduced shared-lock issue locally; it does not establish
production deployment or protection against every form of resource exhaustion.
Code review and commit are pending.

### SEC-02: Persistent tokens lack server-enforced expiration

`internal/app/admin.go:577` validates token hash, revocation and password
fingerprint, but not session age. `internal/household/model.go` stores no session
creation or expiry timestamp. The 400-day cookie expiration (`admin.go:28,603`)
limits browser retention, not replay of a copied token in a manually supplied
Cookie header. A stolen token remains usable until logout/revocation or an
observed password change. Secure/HttpOnly cookies and hashed token storage
reduce theft opportunities; token theft itself was not demonstrated.

**Recommendation:** persist and enforce absolute and idle session deadlines,
offer session inventory/revocation, and test expired-token replay. Consider a
shorter remembered-session lifetime. Existing logout and password-change
invalidation tests passed; those controls do not supply time expiry.

### SEC-03: First-run setup has no independent ownership proof

When no household exists, `internal/app/admin.go:190` exposes setup and
`setup()` at line 381 accepts a valid new household/password without a bootstrap
secret or prior operator identity. Any visitor can obtain their own CSRF nonce;
CSRF is not proof of ownership. A fresh local instance returned HTTP 200 with
the `/setup/finish` form. Existing HTTPS tests confirm that setup creates a
household and privileged session.

This is an intended bootstrap workflow with a deployment-sensitive risk: if a
fresh instance is publicly reachable before its owner initializes it, another
visitor can claim it. It does not bypass authentication on an initialized
household, and no production takeover was attempted.

**Recommendation:** require a single-use operator bootstrap token, or restrict
access at the proxy until setup completes. Document initialization before public
exposure and verify that deleting/loss of persistent storage cannot silently
reopen public enrollment.

### SEC-04: S3 permits cleartext remote transport

`internal/household/s3.go:70-72` accepts both HTTP and HTTPS for any hostname.
An operator can therefore configure a remote HTTP endpoint, exposing household
data and signed requests to network interception/modification. AWS signing is
not encryption; this does not imply that the secret signing key is transmitted
in plaintext. HTTPS is supported and recommended in the README. No insecure
production endpoint was observed and no network interception was attempted.

**Recommendation:** require HTTPS by default; allow HTTP only through an
explicit development override or narrowly scoped local-endpoint policy. Test
rejection of remote cleartext endpoints and retain certificate verification.

### SEC-05: CSP does not restrict script execution

`internal/app/admin.go:144` sends only `frame-ancestors 'none'`, confirmed by a
local GET. This prevents framing but supplies no `script-src`, `object-src`, or
`base-uri` restrictions. No executable XSS was identified: Go `html/template`
provides contextual escaping, and reviewed wish links are scheme-validated.
This is defense-in-depth hardening, not evidence of an existing XSS exploit.

**Recommendation:** introduce a tested script policy using external scripts or
nonces/hashes as needed, plus `object-src 'none'` and `base-uri 'none'`.
Validate appearance and inline-style compatibility before enforcement.

## OWASP Top 10 coverage

| 2025 category | Assessment |
| --- | --- |
| A01 Broken Access Control | Admin gates, explicit link identity, list/wish scope, claimant privacy and CSRF reviewed; related Go/browser regressions passed. No bypass found in reviewed paths. Image-fetch SSRF controls also reviewed: IP filtering, checked DNS pinned to the connection, redirect revalidation. |
| A02 Security Misconfiguration | SEC-03/05; explicit proxy-peer trust, no trust-all CIDRs, secure cookie flags, no-store and nosniff are strengths. Actual Traefik TLS, HSTS, access logs and network isolation were not assessed. |
| A03 Software Supply Chain Failures | Frozen npm lockfile, Go module verification and pinned CI action digests reviewed. Advisory scan results below; registry/publisher compromise and container provenance not independently verified. |
| A04 Cryptographic Failures | Argon2id uses 19 MiB, two iterations, one lane, random 16-byte salt; tokens use 32 random bytes and stored session hashes. SEC-04 concerns transport. Backup/disk/S3 encryption and key management remain deployment checks. |
| A05 Injection | Contextual HTML escaping, HTTP(S) link validation, path-safe IDs, strict model validation and symlink defenses reviewed. No SQL backend or reviewed shell-execution sink in HTTP handling. Encoded asset traversal probe returned 404. No injection exploit established; not exhaustive fuzzing. |
| A06 Insecure Design | SEC-01/03; capability URLs intentionally grant bearer access. Share links should be treated as credentials and redacted from proxy query logs; app logs use fixed operation names. |
| A07 Authentication Failures | SEC-02; password hashing, cookie flags, password-change invalidation, logout, trusted client-IP calculation and login lockout reviewed/tested. Per-IP lockout is in memory and resets on restart by design; distributed attack resistance was not tested. |
| A08 Software or Data Integrity Failures | Atomic writes, revision checks, schema/reference validation and symlink protections reviewed; regression suite passed. Filesystem/S3 write access is a privileged trust boundary; no independent authenticity signature protects data from a compromised storage administrator. |
| A09 Security Logging and Alerting Failures | Fixed operation/error categories avoid leaking content or tokens; failure/logging tests passed. Central retention, alert delivery and incident response were not tested. Successful login/logout lack a distinct audit event beyond generic actions: consider dedicated redacted authentication events. |
| A10 Mishandling of Exceptional Conditions | Failure, read-only cache, revocation and rollback regressions passed. Body/image size limits and server timeouts exist; SEC-01 demonstrates a remaining availability gap. No broad fuzzing or resource-exhaustion campaign performed. |

For OWASP 2021 readers, SSRF corresponds to A10:2021, injection to A03:2021,
and vulnerable/outdated components to A06:2021. This report uses the 2025 category
names rather than treating the 2021 numbering as current.

## Executed checks

Commands ran from `/workspace/wishmas` with Go **1.27.3** and the prepared
toolchain/cache environment. No application source, tests, lockfiles or
dependency declarations were changed.

| Check | Outcome |
| --- | --- |
| `scripts/check-go.sh` | Passed: module download/verification, tidy diff, vet, race-enabled tests across all four packages. |
| `python3 -B -m unittest discover -s tests/release -v` | 20 passed; publication tests use local disposable Git remotes. No actual publication performed. |
| `npm run test:e2e` | 22 passed using isolated HTTPS servers, including cross-origin rejection, login lockout, password-change revocation, share identity and stale-action scenarios. |
| `npm audit --json` | Zero reported advisories at assessment time; development/build dependencies included. |
| `govulncheck -show verbose ./...` (scanner v1.9.0) | No vulnerabilities found in imported packages or reachable application code. |
| Bounded local HTTP probes | Public first-run form 200; encoded asset traversal 404; framing/nosniff/referrer headers present; SEC-01 reproduced and server recovery verified. Disposable process terminated afterward. |

## Limitations and follow-up

Prioritize SEC-01, then server-side session expiry and controlled bootstrap.
Validate proxy buffering/timeouts, HTTPS redirection/HSTS, query-log redaction,
S3 IAM/transport/encryption, backup protection and monitoring in the actual
deployment. S3 coverage here uses fixtures; the opt-in real-provider integration
test was not exercised. No external scan, production TLS audit, full history
secret scan, container image scan, exhaustive endpoint fuzzing or independent
penetration test was performed. Findings remain open until remediated and
retested. This document records assessment results; it does not establish a
disclosure channel or supported-version policy on behalf of the maintainer.
