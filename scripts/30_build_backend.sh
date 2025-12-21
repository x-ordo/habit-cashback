#!/usr/bin/env bash
set -euo pipefail
cd backend
go mod tidy
go build ./...
echo "✅ backend build ok"
