#!/usr/bin/env bash
set -Eeuo pipefail

cat >&2 <<'EOF'
This inherited Docker installer is disabled in DX-3FV.

Use the verified Debian 12 native installer instead:
  bash <(curl -fsSL https://raw.githubusercontent.com/torr9522/dx-3fv/main/scripts/install.sh)

The stable installer selects a fixed published version and verifies its SHA256.
EOF
exit 64
