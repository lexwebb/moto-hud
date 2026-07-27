#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../pi"
go generate ./internal/hudui/...
