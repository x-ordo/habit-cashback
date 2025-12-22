#!/usr/bin/env bash
set -euo pipefail
cd frontend
pnpm i
pnpm build
echo "✅ frontend build ok"
