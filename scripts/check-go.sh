#!/usr/bin/env bash
# @impl AUT-001-AC1 DEP-001-AC3 REL-001-AC2
set -euo pipefail
cd "$(dirname "$0")/.."
export GOFLAGS=-mod=readonly
go mod download
go mod verify
go mod tidy -diff
go vet ./...
# @impl COV-001-AC1: collect coverage in the existing invocation only.
coverage_args=()
if [[ -n "${WISHMAS_COVERAGE_FILE:-}" ]]; then
  coverage_args=(-covermode=atomic "-coverprofile=$WISHMAS_COVERAGE_FILE")
fi
go test -race -count=1 "${coverage_args[@]}" ./...
