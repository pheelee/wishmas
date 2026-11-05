# Security assessment

Current assessment as of **2026-11-05**, using **OWASP Top 10:2025**.

## Overview

**Three findings are mitigated in the current code; two risks are explicitly
accepted by the operator.** Production deployment and production verification
are outside this assessment.

| Finding | Current disposition | Current control or accepted risk |
| --- | --- | --- |
| [SEC-01: Request-body availability](#sec-01-request-body-availability) | Mitigated | Bounded POST parsing occurs outside the shared application lock; a stalled body does not block ordinary pages. |
| [SEC-02: Persistent admin sessions](#sec-02-persistent-admin-sessions) | Risk accepted | Sessions have no server-side time expiry. A stolen token remains usable until revocation or password change. |
| [SEC-03: First-run ownership](#sec-03-first-run-ownership) | Risk accepted | Setup is open until initialization; the administrator is responsible for completing setup on a fresh instance. |
| [SEC-04: S3 transport](#sec-04-s3-transport) | Mitigated | All S3 endpoints require HTTPS, including local/private endpoints. HTTP redirects are rejected and TLS verification remains enabled. |
| [SEC-05: Browser content policy](#sec-05-browser-content-policy) | Mitigated | CSP restricts script sources, embedded objects, base URLs, framing and form destinations; browser tests verify enforcement. |

Risk acceptance is not a technical fix. Reassess these decisions if usage,
exposure or data sensitivity changes. No critical or high-severity vulnerability
has been established within the reviewed scope; this is not a guarantee that
other vulnerabilities do not exist.

## Scope

The review covers Go HTTP handlers, authentication and authorization,
local/S3 persistence, image processing and downloads, embedded HTML/JavaScript,
runtime configuration, and CI security checks. Functional security tests use
isolated local instances and disposable data, including HTTPS browser fixtures.

No public deployment, production household, real S3 account or external target
was tested. This assessment is not a certification or an independent penetration
test.

## Findings and dispositions

### SEC-01: Request-body availability

**Mitigated.** POST request bodies are parsed with a **16 KiB limit before
acquiring the shared household lock**. Parse errors are rendered with the
household locale and existing storage-error precedence. Authentication, CSRF
checks and serialized state mutations remain in place. The server applies a
15-second read timeout.

Regression tests cover stalled bodies before and after setup, malformed and
oversized requests, and independent page/health responses. A local raw HTTP
probe confirms ordinary pages remain available while an incomplete POST is
pending. These controls address shared-lock contention; they do not establish
resistance to every form of resource exhaustion. `/healthz` is a liveness check,
not proof of complete functional readiness.

### SEC-02: Persistent admin sessions

**Risk accepted by the operator.** Admin sessions deliberately have no
server-enforced absolute or idle expiry. The operator accepts this for a family
application with no critical data in their stated usage, prioritizing persistent
logins until logout or a password change.

Tokens contain 256 random bits and are stored as hashes. Browser cookies are
Secure, HttpOnly, host-scoped and SameSite=Lax. Logout/revocation and an observed
password-hash change invalidate sessions, including durable local logout during
S3 outages. The browser cookie has a 400-day retention period, but that does not
limit manual replay of a copied token.

A stolen token can therefore remain valid until revocation or password change.
No token theft was demonstrated. Acceptance does not imply that theft is
impossible or that household information contains no personal data.

### SEC-03: First-run ownership

**Risk accepted by the operator.** Until a household exists, setup remains open
without an independent bootstrap token. The administrator is responsible for
completing setup after starting a fresh instance.

Anyone who can reach an uninitialized instance can choose its initial admin
password. This does not bypass authentication on an initialized household.
The same exposure can arise if missing persistent data causes setup to become
available again. No restricted proxy access is assumed or verified; the
first-visitor ownership window is explicitly accepted.

### SEC-04: S3 transport

**Mitigated.** Environment parsing and direct backend construction require
HTTPS for every S3 endpoint, including loopback and private-network services.
There is **no HTTP opt-in or insecure TLS override**. The client rejects
redirects to HTTP, retains its ten-redirect limit, and verifies TLS certificates
and hostnames. Private certificate authorities must be installed in the runtime
trust store.

Tests verify HTTP rejection before network access, blocked downgrade redirects,
and rejection of an untrusted certificate. S3 protocol/browser fixtures use
HTTPS with explicitly trusted fixture certificates. The default local-storage
mode does not require S3.

Operators must configure a trusted HTTPS endpoint. Changing an endpoint selects
a different local cache/picture/revocation scope; see the
[S3 configuration instructions](docs/s3-storage.md).

### SEC-05: Browser content policy

**Mitigated.** HTML responses enforce the following Content Security Policy:

```text
script-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'
```

The policy permits the same-origin JavaScript bundle and blocks inline scripts,
inline event handlers, eval, external scripts, embedded objects, base URL
changes and cross-origin form submissions. Existing inline appearance styles
remain permitted; this focused policy does not restrict every resource type.
Contextual Go template escaping remains the primary output-encoding control.
No executable XSS was identified.

Browser tests verify actual enforcement and normal bundle/UI operation. Eval is
probed from a real script context rather than privileged DevTools evaluation;
external test URLs are intercepted rather than contacted.

## OWASP Top 10 coverage

| Category | Current assessment |
| --- | --- |
| A01 Broken Access Control | Admin gates, scoped share-link identities, wish/picture access, claimant privacy and CSRF have regression coverage. No bypass was established. Image-download SSRF controls filter addresses, pin checked DNS results to connections and revalidate redirects. |
| A02 Security Misconfiguration | Explicit proxy-peer trust, rejection of trust-all CIDRs, secure cookie flags, no-store/nosniff headers and CSP are implemented. First-run ownership risk is accepted under SEC-03. Production proxy/TLS configuration is unverified. |
| A03 Software Supply Chain Failures | Lockfile/module integrity checks, digest-pinned CI actions, golangci-lint and govulncheck are configured. Registry compromise and container provenance were not independently verified. |
| A04 Cryptographic Failures | Passwords use Argon2id with 19 MiB memory, two iterations, one lane and random 16-byte salts. Session tokens use cryptographic randomness and hashed storage. S3 requires verified HTTPS. Disk, backup and provider-side encryption remain deployment checks. |
| A05 Injection | Contextual template escaping, HTTP(S) wish-link validation, path-safe identifiers, model validation and root-scoped/no-follow file operations protect reviewed paths. No injection exploit was established. There is no SQL backend. |
| A06 Insecure Design | Request-body parsing avoids shared-lock contention. Open first-run setup is an accepted design risk. Share links are bearer credentials and should be redacted from proxy query logs. |
| A07 Authentication Failures | Password checks, cookie flags, logout, password-change invalidation and trusted-client-IP login throttling are tested. Indefinite session lifetime is accepted under SEC-02. In-memory login throttling resets on restart; distributed attacks were not tested. |
| A08 Software or Data Integrity Failures | Atomic writes, revision checks, model/reference validation and filesystem protections have regression coverage. Filesystem/S3 administrators remain trusted; stored household data has no independent authenticity signature. |
| A09 Security Logging and Alerting Failures | Fixed operation names and bounded error categories avoid raw credential/content/URL logging. Write and cleanup failures have diagnostics. Dedicated login/logout audit events, centralized retention and alerting are not established by this assessment. |
| A10 Mishandling of Exceptional Conditions | Body/image limits, timeouts, rollback, cache fallback, revocation, cleanup failures and response-write failures have regression coverage. Broad fuzzing and sustained resource-exhaustion testing were not performed. |

## Verification evidence

Recorded verification results for the current implementation:

| Check | Recorded outcome |
| --- | --- |
| `golangci-lint config verify` and `golangci-lint run ./...` | Passed with zero findings. Enabled linters: gosec, govet, ineffassign, staticcheck, errorlint, bodyclose and nilerr. One narrowly justified healthcheck G704 suppression is retained. |
| `scripts/check-go.sh` | Passed: module verification, tidy diff, vet and race-enabled application tests. |
| `CGO_ENABLED=0 go build -tags nodynamic ./...` | Passed. |
| `govulncheck ./...` | No vulnerable imported packages or reachable application code reported. |
| `npm audit --json` | Zero reported advisories at the time checked. |
| Automation/release tests | 20 passed. |
| `npm run test:e2e` | 24 passed, including CSP enforcement and HTTPS-backed S3 scenarios. |
| Targeted security tests | Stalled POST availability, malformed/oversized requests, TLS/downgrade rejection, file/symlink protections and redacted error handling verified. |

Toolchain: Go **1.27.3**, golangci-lint **v2.15.0**, govulncheck **v1.9.0**.
The CI workflow includes pinned lint/vulnerability tools; contributor instructions
require local lint before commits. Hosted execution of the updated workflow is
not verified here. Passing scans are bounded by tool coverage and advisory data.

## Remaining responsibilities and limitations

- Deploy and verify the mitigations in the actual environment. Local validation
  does not establish production protection.
- The administrator owns fresh-instance initialization and accepts persistent
  sessions as described above. Revisit acceptance if the app's use changes.
- Verify proxy buffering/timeouts, HTTPS redirection/HSTS, query-log redaction,
  S3 permissions/encryption, backup protection and monitoring in deployment.
- S3 tests use fixtures; the opt-in real-provider integration was not exercised.
- No public scan, production TLS audit, full-history secret scan, container-image
  scan, exhaustive fuzzing or independent penetration test was performed.

This document does not establish a disclosure channel or supported-version
policy on behalf of the maintainer.
