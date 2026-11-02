# Dependency baseline

Latest stable versions were resolved on **2026-11-01** for the V1 preparation. This is a recorded baseline, not a claim that versions remain latest after that date. Exact application dependencies and integrity records are in go.mod/go.sum and package.json/package-lock.json.

| Dependency | Previous | Verified stable baseline |
| --- | --- | --- |
| Go toolchain / builder / container helper | 1.27.3 | 1.27.3 (already latest stable) |
| AWS SDK core | 1.47.4 | 1.47.5 |
| AWS credentials | 1.20.10 | 1.20.13 |
| AWS S3 | 1.114.5 | 1.114.6 |
| AWS Smithy | 1.28.6 | 1.28.7 |
| YAML v3 | 3.0.5 | 3.0.6 |
| purego | 0.10.1 | 0.11.2 |
| Tailwind CSS / CLI | 4.3.0 | 4.3.3 |
| Playwright | 1.64.1 | 1.64.1 (already current) |
| AVIF / WebP codecs | 0.6.0 / 0.6.4 | unchanged, current stable |
| Traefik container test proxy (removed 2026-11-02) | previous digest | Historical: 3.7.17; no longer a test dependency |

AWS indirect modules were updated consistently. Other selected direct/indirect application modules were already current. `go get -u ./...`, `go get -u all`, tidy, module verification and npm's latest dist-tags/update were used. YAML v4 was checked and remains a release candidate (`v4.0.0-rc.7`), so it was not adopted. Major upgrades were allowed; no stable major migration was identified for the selected application packages.

The full upstream Go module graph can include modules not imported by this application or its tests. `golang.org/x/net` remains at its upstream graph version (v0.60.0 versus available v0.62.0); `go mod why -m` confirms the main module does not need it. No artificial unused requirement was added just to override upstream metadata. It is not an omitted imported runtime dependency. Re-evaluate if a future application dependency imports it.

The frontend/package baseline is **0.1.0**, consistently in both npm records and the root lockfile package. No release/tag was published by this refresh.

Verification passed: full Go race tests, vet, module verification/tidy and static Go-only build; frontend rebuild; all 22 Chromium scenarios; English/German real Docker/Traefik HTTPS smoke and Compose checks. The initial browser run passed 20 and failed two before page execution because cold Go compilation exceeded the fixture's 30-second limit. The binary-build fixture now has a bounded 180-second timeout; a clean full rerun passed all 22 with the same page assertions. Original household data and live apps were not changed.
