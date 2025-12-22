#!/usr/bin/env bash
set -euo pipefail

echo "[1/3] iframe check (YouTube only exception) ..."
if command -v rg >/dev/null 2>&1; then
  if rg -n "<iframe" frontend/src; then
    echo "Found iframe usage. Remove it unless it's YouTube embed only."
    exit 1
  fi
else
  echo "ripgrep not installed. Skipping iframe scan."
fi

echo "[2/3] ensure legal routes exist ..."
grep -q 'path="/terms"' frontend/src/app/App.tsx
grep -q 'path="/privacy"' frontend/src/app/App.tsx
grep -q 'path="/support"' frontend/src/app/App.tsx
echo "OK"

echo "[3/3] done"
