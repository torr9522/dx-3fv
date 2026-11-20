#!/usr/bin/env bash
set -euo pipefail

release_dir=${1:?usage: install-agent-native.sh RELEASE_DIR AGENT_CONFIG}
agent_config=${2:?usage: install-agent-native.sh RELEASE_DIR AGENT_CONFIG}
release_dir=$(realpath -- "$release_dir")
agent_config=$(realpath -- "$agent_config")
test -x "$release_dir/bin/dx-3fv-agent"
test -f "$release_dir/MANIFEST"
test -f "$release_dir/deploy/agent-gost.json"
test -f "$agent_config"
ca_bundle=${SSL_CERT_FILE:-/etc/ssl/certs/ca-certificates.crt}
if [[ ! -s "$ca_bundle" ]]; then
  printf 'DX-3FV Agent requires a non-empty CA bundle: %s\n' "$ca_bundle" >&2
  exit 1
fi
(cd "$release_dir" && sha256sum -c SHA256SUMS)
version=$(sed -n 's/^VERSION=//p' "$release_dir/MANIFEST")
test -n "$version"

install -d -m 0755 /opt/dx-3fv-agent/releases /etc/dx-3fv-agent
target="/opt/dx-3fv-agent/releases/dx-3fv-agent-$version"
test ! -e "$target"
install -d -m 0755 "$target"
install -m 0755 "$release_dir/bin/dx-3fv-agent" "$target/dx-3fv-agent"
install -m 0600 "$agent_config" "$target/config.json"
install -m 0600 "$release_dir/deploy/agent-gost.json" "$target/gost.json"
install -m 0644 "$release_dir/deploy/dx-3fv-agent.service" /etc/systemd/system/dx-3fv-agent.service
ln -sfn "$target" /opt/dx-3fv-agent/current
systemctl daemon-reload
systemctl enable --now dx-3fv-agent.service
ready_checks=0
for _ in $(seq 1 30); do
  active_state=$(systemctl is-active dx-3fv-agent.service 2>/dev/null || true)
  sub_state=$(systemctl show -p SubState --value dx-3fv-agent.service 2>/dev/null || true)
  if [[ "$active_state" == "active" && "$sub_state" == "running" ]]; then
    ((ready_checks += 1))
    if [[ "$ready_checks" -ge 3 ]]; then
      break
    fi
  else
    ready_checks=0
  fi
  sleep 1
done
test "$ready_checks" -ge 3
