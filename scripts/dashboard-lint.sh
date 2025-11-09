#!/bin/bash
set -euo pipefail

# Navigate to the dashboard directory where package.json exists
dirs_to_lint="dashboard"

for dir in $dirs_to_lint; do
  pushd "$dir"

  # Ensure Yarn (v4) is available
  if ! command -v yarn >/dev/null 2>&1; then
    if command -v corepack >/dev/null 2>&1; then
      corepack enable >/dev/null 2>&1 || true
      # Use the version pinned in package.json's "packageManager"
      corepack install -g >/dev/null 2>&1 || true
      corepack prepare yarn@4.9.2 --activate >/dev/null 2>&1 || true
    else
      # Fallback: install corepack in the node environment used by pre-commit
      if command -v npm >/dev/null 2>&1; then
        npm install -g corepack >/dev/null 2>&1 || true
        corepack enable >/dev/null 2>&1 || true
        corepack prepare yarn@4.9.2 --activate >/dev/null 2>&1 || true
      fi
    fi
  fi

  # As a last resort, try npx to run corepack without global install
  if ! command -v yarn >/dev/null 2>&1 && command -v npx >/dev/null 2>&1; then
    npx --yes corepack enable >/dev/null 2>&1 || true
    npx --yes corepack prepare yarn@4.9.2 --activate >/dev/null 2>&1 || true
  fi

  if ! command -v yarn >/dev/null 2>&1; then
    echo "Error: yarn is still unavailable. Ensure Node.js is installed and corepack is enabled." >&2
    exit 127
  fi

  # Install deps to create the Yarn install state, then run lint
  yarn install --immutable
  yarn lint
  popd
done
