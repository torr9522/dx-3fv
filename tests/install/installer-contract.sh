#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
# shellcheck source=../../scripts/install.sh
. "$repo_root/scripts/install.sh"

validate_domain panel.example.com
! validate_domain 'https://panel.example.com'
! validate_domain 'panel_example.com'
validate_username 'dxadmin_1234'
! validate_username 'bad user'
validate_password 'Strong-Password-2026!'
! validate_password 'short'
[[ $(artifact_name v1.0.0-rc.3) == dx-3fv-v1.0.0-rc.3.tar.gz ]]
[[ $(artifact_sha256 v1.0.0-rc.3) == 85d5aa4e5e68b2d286bfae08efc97a0a3e16febaa2684ff8226c95c43b59874f ]]
! artifact_name latest

archive="$repo_root/release/artifacts/$(artifact_name v1.0.0-rc.3)"
printf '%s  %s\n' "$(artifact_sha256 v1.0.0-rc.3)" "$archive" | sha256sum -c -
tar -tzf "$archive" 'dx-3fv-v1.0.0-rc.3/scripts/install-native.sh' >/dev/null

echo 'installer contract: PASS'
