#!/usr/bin/env bash
set -euo pipefail

release_dir=${1:?usage: upgrade-native.sh RELEASE_DIR}
release_dir=$(realpath -- "$release_dir")
test -x "$release_dir/bin/dx-3fv-paneld"
(cd "$release_dir" && sha256sum -c SHA256SUMS)
version=$(sed -n 's/^VERSION=//p' "$release_dir/MANIFEST")
test -n "$version"
target="/opt/dx-3fv/releases/dx-3fv-$version"
test ! -e "$target"
previous=$(readlink -f /opt/dx-3fv/current)
test -n "$previous"
stopped=0
env_changed=0
env_backup=""
restart_previous_on_error() {
  if [[ "$stopped" == 1 ]]; then
    ln -sfn "$previous" /opt/dx-3fv/current
    if [[ "$env_changed" == 1 && -n "$env_backup" ]]; then
      install -m 0600 -o root -g root "$env_backup" /etc/dx-3fv/panel.env
    fi
    systemctl start dx-3fv-panel.service || true
  fi
}
trap restart_previous_on_error ERR

backup_dir="/var/lib/dx-3fv/backups/$(date -u +%Y%m%dT%H%M%SZ)"
install -d -m 0750 -o dx-3fv -g dx-3fv "$backup_dir"
env_backup="$backup_dir/panel.env"
install -m 0600 -o root -g root /etc/dx-3fv/panel.env "$env_backup"
systemctl stop dx-3fv-panel.service
stopped=1
if [[ -f /var/lib/dx-3fv/dx-3fv.db ]]; then
  sqlite3 /var/lib/dx-3fv/dx-3fv.db ".backup '$backup_dir/dx-3fv.db'"
  chown dx-3fv:dx-3fv "$backup_dir/dx-3fv.db"
fi
cp -a "$release_dir" "$target"
chown -R root:root "$target"
sed -i "s/^FLUX_VERSION=.*/FLUX_VERSION=$version/" /etc/dx-3fv/panel.env
env_changed=1
ln -sfn "$target" /opt/dx-3fv/current
systemctl start dx-3fv-panel.service
if ! "$target/scripts/health-check.sh"; then
  systemctl stop dx-3fv-panel.service
  ln -sfn "$previous" /opt/dx-3fv/current
  if [[ -f "$backup_dir/dx-3fv.db" ]]; then
    install -m 0640 -o dx-3fv -g dx-3fv "$backup_dir/dx-3fv.db" /var/lib/dx-3fv/dx-3fv.db
  fi
  install -m 0600 -o root -g root "$env_backup" /etc/dx-3fv/panel.env
  env_changed=0
  systemctl start dx-3fv-panel.service
  stopped=0
  "$previous/scripts/health-check.sh"
  exit 1
fi
stopped=0
env_changed=0
trap - ERR
printf '%s\n' "$previous" >/var/lib/dx-3fv/previous-release
chown dx-3fv:dx-3fv /var/lib/dx-3fv/previous-release
