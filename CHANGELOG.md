# Changelog

All notable changes to this project will be documented in this file.

Please choose versions by [Semantic Versioning](http://semver.org/).

* MAJOR version when you make incompatible API changes,
* MINOR version when you add functionality in a backwards-compatible manner, and
* PATCH version when you make backwards-compatible bug fixes.

## Unreleased

- feat: add a cluster build arm so the image can be built in Kubernetes, not only on the Mac

## v0.1.6

- chore(dark-factory): worktree workflow with PR delivery

## v0.1.5

- chore: update github.com/bborbe/collection to v1.21.6, github.com/bborbe/cqrs to v0.6.15, github.com/bborbe/errors to v1.6.6, github.com/bborbe/http to v1.26.32, github.com/bborbe/kafka to v1.29.11, github.com/bborbe/kv to v1.21.28, github.com/bborbe/notification to v0.7.10, github.com/bborbe/run to v1.11.10, github.com/bborbe/sentry to v1.10.7, github.com/bborbe/service to v1.11.2, github.com/bborbe/time to v1.27.25, github.com/bborbe/validation to v1.5.8, github.com/onsi/ginkgo/v2 to v2.34.0

## v0.1.4
- chore: bump Go toolchain directive to 1.27.3

## v0.1.3

- fix: bump `osv-scanner` to v2.6.0 and `golang.org/x/net` to v0.60.0 so the Linux vulnerability gates stop failing. v2.3.1 pins `golang.org/x/tools` v0.38.0, whose SSA builder aborts with `unexpected expr: *ast.KeyValueExpr` on the promoted-field composite-literal key Go 1.27 permits in the Linux stdlib, so a repo on the old pin passes locally on darwin and fails only in Linux CI. `x/net` v0.58.0 carries `GO-2026-6603/6610/6611/6612/6617`, which fail both `vulncheck` and `trivy`.

## v0.1.2

- fix: rebuild image from post-bump tree (v0.1.1 image was built from pre-bump code)

## v0.1.1

- bump github.com/bborbe/notification to v0.2.0 (adds go-release notification type support)

## v0.1.0

- feat: extract notification discord service from trading monorepo (sends notifications to Discord channels)
