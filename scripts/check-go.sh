#!/usr/bin/env bash
# @impl AUT-001-AC1 DEP-001-AC3 REL-001-AC2
set -euo pipefail
cd "$(dirname "$0")/.."
export GOFLAGS=-mod=readonly
go mod download
go mod verify
go mod tidy -diff
go vet ./...
go test -race -count=1 ./...
