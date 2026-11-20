#!/usr/bin/env bash
set -euo pipefail

target=${1:-}
database_backup=${2:-}
if [[ -z "$target" && -f /var/lib/dx-3fv/previous-release ]]; then
  target=$(cat /var/lib/dx-3fv/previous-release)
fi
test -n "$target"
target=$(realpath -- "$target")
test -x "$target/bin/dx-3fv-paneld"
target_version=$(sed -n 's/^VERSION=//p' "$target/MANIFEST")
test -n "$target_version"
current=$(readlink -f /opt/dx-3fv/current)
systemctl stop dx-3fv-panel.service
rollback_safety_db=""
rollback_safety_env=$(mktemp /etc/dx-3fv/panel.env.rollback.XXXXXX)
install -m 0600 -o root -g root /etc/dx-3fv/panel.env "$rollback_safety_env"
if [[ -n "$database_backup" ]]; then
  database_backup=$(realpath -- "$database_backup")
  test -f "$database_backup"
  if [[ -f /var/lib/dx-3fv/dx-3fv.db ]]; then
    rollback_safety_db=$(mktemp /var/lib/dx-3fv/rollback-safety.XXXXXX.db)
    sqlite3 /var/lib/dx-3fv/dx-3fv.db ".backup '$rollback_safety_db'"
    chown dx-3fv:dx-3fv "$rollback_safety_db"
  fi
  install -m 0640 -o dx-3fv -g dx-3fv "$database_backup" /var/lib/dx-3fv/dx-3fv.db
fi
sed -i "s/^FLUX_VERSION=.*/FLUX_VERSION=$target_version/" /etc/dx-3fv/panel.env
ln -sfn "$target" /opt/dx-3fv/current
systemctl start dx-3fv-panel.service
if ! "$target/scripts/health-check.sh"; then
  systemctl stop dx-3fv-panel.service
  ln -sfn "$current" /opt/dx-3fv/current
  if [[ -n "$rollback_safety_db" ]]; then
    install -m 0640 -o dx-3fv -g dx-3fv "$rollback_safety_db" /var/lib/dx-3fv/dx-3fv.db
  fi
  install -m 0600 -o root -g root "$rollback_safety_env" /etc/dx-3fv/panel.env
  systemctl start dx-3fv-panel.service
  rm -f -- "$rollback_safety_db"
  rm -f -- "$rollback_safety_env"
  exit 1
fi
if [[ -n "$rollback_safety_db" ]]; then
  rm -f -- "$rollback_safety_db"
fi
rm -f -- "$rollback_safety_env"
printf '%s\n' "$current" >/var/lib/dx-3fv/previous-release
