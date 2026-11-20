#!/usr/bin/env bash
set -euo pipefail

release_dir=${1:?usage: install-native.sh RELEASE_DIR}
release_dir=$(realpath -- "$release_dir")
test -x "$release_dir/bin/dx-3fv-paneld"
test -f "$release_dir/MANIFEST"
(cd "$release_dir" && sha256sum -c SHA256SUMS)
version=$(sed -n 's/^VERSION=//p' "$release_dir/MANIFEST")
test -n "$version"

if ! id dx-3fv >/dev/null 2>&1; then
  useradd --system --home /var/lib/dx-3fv --shell /usr/sbin/nologin dx-3fv
fi
install -d -m 0755 /opt/dx-3fv/releases /etc/dx-3fv
install -d -m 0750 -o dx-3fv -g dx-3fv /var/lib/dx-3fv /var/log/dx-3fv
target="/opt/dx-3fv/releases/dx-3fv-$version"
test ! -e "$target"
cp -a "$release_dir" "$target"
chown -R root:root "$target"

if [[ ! -f /etc/dx-3fv/panel.env ]]; then
  umask 077
  jwt_secret=$(openssl rand -hex 32)
  sed -e "s/replace-on-target/$jwt_secret/" -e "s/replace-with-release-version/$version/" \
    "$target/deploy/panel.env.example" >/etc/dx-3fv/panel.env
fi
install -m 0644 "$target/deploy/dx-3fv-panel.service" /etc/systemd/system/dx-3fv-panel.service
ln -sfn "$target" /opt/dx-3fv/current
systemctl daemon-reload
systemctl enable --now dx-3fv-panel.service
if ! "$target/scripts/health-check.sh"; then
  systemctl disable --now dx-3fv-panel.service || true
  rm -f /opt/dx-3fv/current
  exit 1
fi
