#!/usr/bin/env bash
set -euo pipefail

for _ in $(seq 1 30); do
  if systemctl is-active --quiet dx-3fv-panel.service && curl -fsS --max-time 2 http://127.0.0.1:17365/flow/test >/dev/null; then
    exit 0
  fi
  sleep 1
done
systemctl status dx-3fv-panel.service --no-pager -l >&2 || true
exit 1
