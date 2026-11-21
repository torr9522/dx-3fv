#!/usr/bin/env bash
set -Eeuo pipefail

cat >&2 <<'EOF'
This inherited FLVX agent installer is disabled in DX-3FV.

DX-3FV agents must be installed from the verified native release package with:
  scripts/install-agent-native.sh

Release source: https://github.com/torr9522/dx-3fv/releases
EOF
exit 64
