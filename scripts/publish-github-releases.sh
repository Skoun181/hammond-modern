#!/usr/bin/env bash
# Publish versioned GitHub releases for known tested Hammond Modern commits.
# Run manually from a Git clone after installing/authenticating GitHub CLI.
set -euo pipefail

REPO="Skoun181/hammond-modern"
STABLE_SHA="5477bb1a5199f44f6ca3ddf41b6a171344b18d9d"
DEVELOP_SHA="db88b500f9cd07ca3d20646e1b4ee11713cd1bf4"

cd "$(dirname "$0")/../.."
command -v gh >/dev/null || { echo "GitHub CLI fehlt (sudo apt install gh)"; exit 1; }
gh auth status >/dev/null || { echo "Bitte zuerst 'gh auth login' ausführen."; exit 1; }
command -v git >/dev/null || { echo "git fehlt."; exit 1; }

verify_branch() {
  local branch="$1" expected="$2" actual
  actual="$(git ls-remote origin "refs/heads/$branch" | awk '{print $1}')"
  if [[ "$actual" != "$expected" ]]; then
    echo "STOPP: $branch zeigt auf ${actual:-<nicht vorhanden>}, erwartet $expected"
    echo "Keine Änderungen durchgeführt. Bitte Branch-Stand prüfen."
    exit 1
  fi
}
verify_branch "stable/hammond-modern" "$STABLE_SHA"
# Develop now includes documentation-only release notes, so verify the tested
# commit remains in its ancestry rather than requiring an unchanged HEAD.
if ! git merge-base --is-ancestor "$DEVELOP_SHA" "origin/develop/hammond-modern-next"; then
  git fetch origin develop/hammond-modern-next
  git merge-base --is-ancestor "$DEVELOP_SHA" "origin/develop/hammond-modern-next" || {
    echo "STOPP: getesteter Develop-Commit ist nicht mehr in der Branch-Historie."; exit 1;
  }
fi

for version in v1.0.0 v1.1.0; do
  if [[ -n "$(git ls-remote origin "refs/tags/$version")" ]]; then
    echo "STOPP: Tag $version existiert bereits; bitte zuerst prüfen."
    exit 1
  fi
done

echo "Release-Ziele:"
echo "  v1.0.0 -> $STABLE_SHA"
echo "  v1.1.0 -> $DEVELOP_SHA"
echo "  Standardbranch -> stable/hammond-modern"
echo "  master wird nach erfolgreicher Release-Erstellung gelöscht."
read -r -p "Alle Schritte wirklich durchführen? (ja/NEIN): " answer
[[ "$answer" == "ja" ]] || { echo "Abgebrochen."; exit 0; }

gh release create v1.0.0 --repo "$REPO" --target "$STABLE_SHA" \
  --title "Hammond Modern v1.0.0" --notes-file docs/releases/v1.0.0.md

gh release create v1.1.0 --repo "$REPO" --target "$DEVELOP_SHA" \
  --title "Hammond Modern v1.1.0" --notes-file docs/releases/v1.1.0.md

gh api --method PATCH "repos/$REPO" -f default_branch=stable/hammond-modern >/dev/null

actual_default="$(gh api "repos/$REPO" --jq '.default_branch')"
[[ "$actual_default" == "stable/hammond-modern" ]] || {
  echo "STOPP: Standardbranch konnte nicht bestätigt werden. master bleibt erhalten."; exit 1;
}
gh api --method DELETE "repos/$REPO/git/refs/heads/master" >/dev/null

echo "FERTIG: Releases v1.0.0 / v1.1.0, Standardbranch stable/hammond-modern, master gelöscht."
