#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   OWNER="x-ordo" REPO="habit-cashback" bash scripts/30_update_branding_en.sh
# Optional:
#   HOMEPAGE_URL="https://..." (if you want repo homepage)
#   APPLY_README_PATCH="1" (if you want sed-based README patch)

: "${OWNER:?OWNER required}"
: "${REPO:?REPO required}"

FULL="$OWNER/$REPO"

# 1) Repo description (+ homepage optional)
DESCRIPTION="습관환급 | Habit Cashback — 습관을 지키면 환급(리워드)되는 Apps in Toss 미션 플랫폼 (Go backend)"

if [[ -n "${HOMEPAGE_URL:-}" ]]; then
  gh repo edit "$FULL" --description "$DESCRIPTION" --homepage "$HOMEPAGE_URL"
else
  gh repo edit "$FULL" --description "$DESCRIPTION"
fi

# 2) Repo variables (display names)
gh variable set SERVICE_DISPLAY_NAME_KO -R "$FULL" --body "습관환급" >/dev/null
gh variable set SERVICE_DISPLAY_NAME_EN -R "$FULL" --body "Habit Cashback" >/dev/null

# Keep existing internal slug (don't change if already set); if missing, set it.
if ! gh variable list -R "$FULL" | grep -q '^SERVICE_NAME'; then
  gh variable set SERVICE_NAME -R "$FULL" --body "habitcashback" >/dev/null
fi

echo "OK: repo + variables updated"
echo " - SERVICE_DISPLAY_NAME_KO=습관환급"
echo " - SERVICE_DISPLAY_NAME_EN=Habit Cashback"

# 3) Optional: patch README first heading line (local repo file)
if [[ "${APPLY_README_PATCH:-}" == "1" ]]; then
  if [[ ! -f README.md ]]; then
    echo "Skip README patch: README.md not found in current directory."
    exit 0
  fi

  if grep -q '습관환급 | Habit Cashback' README.md; then
    echo "README already has bilingual title."
    exit 0
  fi

  if head -n 1 README.md | grep -q '^#'; then
    sed -i '' '1s/^# .*/# 습관환급 | Habit Cashback/' README.md
  else
    (echo "# 습관환급 | Habit Cashback"; echo ""; cat README.md) > README.md.tmp && mv README.md.tmp README.md
  fi

  echo "README patched. Commit and push to apply."
fi
