#!/usr/bin/env bash
# Refresh the vendored 40kdc-data snapshot under data/40kdc.
#
# Copies only the entities the app models (factions, detachments, stratagems,
# force dispositions, missions, mission matchups, secondary/primary cards,
# deployment patterns). Army-list/unit/weapon/terrain data is intentionally NOT
# vendored. Run from anywhere; paths are resolved relative to the repo root.
#
# Usage:
#   scripts/vendor-40kdc-data.sh                 # clone upstream main and vendor
#   scripts/vendor-40kdc-data.sh /path/to/40kdc-data   # vendor from a local checkout
set -euo pipefail

REPO_URL="https://github.com/wn-mitch/40kdc-data"
BRANCH="main"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DST="$REPO_ROOT/data/40kdc"

cleanup() { [ -n "${TMP:-}" ] && rm -rf "$TMP"; }
trap cleanup EXIT

if [ "$#" -ge 1 ]; then
  SRC_REPO="$1"
else
  TMP="$(mktemp -d)"
  echo "Cloning $REPO_URL ($BRANCH)..."
  git clone --depth 1 --branch "$BRANCH" "$REPO_URL" "$TMP"
  SRC_REPO="$TMP"
fi

COMMIT="$(git -C "$SRC_REPO" rev-parse HEAD)"
COMMIT_DATE="$(git -C "$SRC_REPO" show -s --format=%cI HEAD)"
SRC="$SRC_REPO/data/core"
VENDORED_AT="$(date -u +%Y-%m-%d)"

if [ ! -d "$SRC" ]; then
  echo "error: $SRC not found — is this a 40kdc-data checkout?" >&2
  exit 1
fi

echo "Vendoring from commit $COMMIT ($COMMIT_DATE)..."
rm -rf "$DST"
mkdir -p "$DST/factions"

# Top-level shared data.
for f in force-dispositions missions mission-matchups secondary-cards deployment-patterns stratagems game-versions; do
  cp "$SRC/$f.json" "$DST/$f.json"
done

# Per-faction: only the entities the app consumes.
for d in "$SRC"/*/; do
  name="$(basename "$d")"
  case "$name" in _example | _reports) continue ;; esac
  [ -f "$d/factions.json" ] || continue
  mkdir -p "$DST/factions/$name"
  for f in factions detachments stratagems; do
    [ -f "$d/$f.json" ] && cp "$d/$f.json" "$DST/factions/$name/$f.json"
  done
done

cat >"$DST/SOURCE.json" <<EOF
{
  "_comment": "Vendored snapshot of reference data from 40kdc-data. Do not edit by hand — refreshed by scripts/vendor-40kdc-data.sh (driven by .github/workflows/update-40kdc-data.yml). App consumes only the entities it models; army-list/unit/weapon/terrain data is intentionally NOT vendored.",
  "repo": "$REPO_URL",
  "branch": "$BRANCH",
  "commit": "$COMMIT",
  "commit_date": "$COMMIT_DATE",
  "edition": "11th",
  "dataslate": "launch",
  "vendored_at": "$VENDORED_AT"
}
EOF

echo "Vendored $(find "$DST" -name '*.json' | wc -l | tr -d ' ') files into $DST"
