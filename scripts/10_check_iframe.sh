#!/usr/bin/env bash
set -euo pipefail
echo "[check] iframe usage (allowed only for YouTube embeds)"
if grep -RIn --line-number "<iframe" frontend/src >/dev/null 2>&1; then
  echo "❌ iframe found:"
  grep -RIn --line-number "<iframe" frontend/src
  exit 1
fi
echo "✅ no iframe found"
