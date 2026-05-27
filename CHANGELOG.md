# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [Unreleased]

### Security

- Upgraded `golang.org/x/net` from `v0.0.0-20210226...` → `v0.55.0`, resolving CVEs GO-2026-5026, GO-2026-4918, GO-2026-4440/4441, and associated HTML parser findings (GO-2026-5024 through 5030).
- Upgraded `golang.org/x/sys` to `v0.45.0` and `golang.org/x/text` to `v0.37.0`.
- `google.golang.org/grpc` resolved to `v1.81.1` via MVS, fixing Symbol-level CVEs GO-2026-4762 (authorization bypass) and GO-2023-2153 (HTTP/2 Rapid Reset DoS). Both were reachable through the binary's call graph.
- `google.golang.org/protobuf` resolved to `v1.36.11`, fixing GO-2024-2611 (infinite loop in JSON unmarshaling).
- `govulncheck ./...` now reports **no vulnerabilities**.

### Added

- `golang.org/x/vuln/cmd/govulncheck` added as a Go tool dependency (`go.mod` `tool` directive).
- `TestNoSymbolVulnerabilities` regression test in `plugin/plugin_test.go`: shells out to `govulncheck -json` and fails if any Symbol-level CVE is reachable from the binary's call graph. Catches future dependency changes that re-introduce runtime-reachable vulnerabilities.
- `TestApmPlugin_SetConfig` table-driven tests covering: default region, explicit region override, static credentials, static credentials with session token, partial credentials fallback, and combined region + credentials.
- `TestApmPlugin_PluginInfo` test asserting plugin name is set correctly.

### Changed

- Go toolchain bumped from `1.23.7` → `1.25.3` (`.go-version` and both CI/release workflows).
- `actions/setup-go` bumped from `v2` → `v5` in `.github/workflows/ci.yml` and `.github/workflows/release.yml`.
- `actions/checkout` bumped from `v2` → `v4` in `.github/workflows/release.yml`.
- `.gitignore` expanded with the canonical Go and macOS templates from `github/gitignore`, covering build artifacts, test binaries, coverage profiles, `go.work`, `.env`, `.DS_Store`, and macOS system directories.

---

## [0.0.3] — Prior release

### Changed

- Updated AWS SDK dependencies (`aws-sdk-go-v2`).
- Plugin now uses default AWS credential chain when `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` environment variables are not set.

---

## [0.0.2] — Prior release

### Added

- Initial open-source release of the Nomad Autoscaler CloudWatch APM plugin.
- Plugin configuration support for `aws_region`, `aws_access_key_id`, `aws_secret_access_key`, and `aws_session_token`.
- `Query` and `QueryMultiple` methods implementing the Nomad Autoscaler APM interface.
- Vagrant example demonstrating the plugin with a working Nomad + CloudWatch setup.
