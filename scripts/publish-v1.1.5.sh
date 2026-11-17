#!/usr/bin/env bash
# Create the v1.1.5 GitHub release AFTER the stable branch has been updated.
# Requires GitHub CLI (gh) and an authenticated GitHub account with release rights.
set -euo pipefail
cd "$(dirname "$0")/.."

repo="Skoun181/hammond-modern"
git fetch origin stable/hammond-modern develop/hammond-modern-next --tags

stable_sha="$(git rev-parse refs/remotes/origin/stable/hammond-modern)"
develop_sha="$(git rev-parse refs/remotes/origin/develop/hammond-modern-next)"

if [[ "$stable_sha" != "$develop_sha" ]]; then
  echo "STOPP: Stable und Develop haben unterschiedliche Commits."
  echo "Stable:  $stable_sha"
  echo "Develop: $develop_sha"
  echo "Bitte erst die vorgesehene Version auf Stable setzen."
  exit 1
fi

gh auth status >/dev/null
if gh release view v1.1.5 --repo "$repo" >/dev/null 2>&1; then
  echo "Release v1.1.5 existiert bereits, kein zweiter Release erstellt."
  exit 0
fi

echo "GitHub Release v1.1.5 wird für Commit $stable_sha erstellt."
read -r -p "Veröffentlichen? Tippe ja: " answer
[[ "$answer" == "ja" ]] || { echo "Abgebrochen."; exit 0; }

gh release create v1.1.5 \
  --repo "$repo" \
  --target "$stable_sha" \
  --title "Hammond Modern v1.1.5" \
  --notes-file docs/releases/v1.1.5.md \
  --latest

echo "Release: https://github.com/$repo/releases/tag/v1.1.5"
