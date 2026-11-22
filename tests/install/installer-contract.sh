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
[[ $DEFAULT_VERSION == v1.0.6 ]]
[[ $INSTALLER_VERSION == 1.0.6 ]]
[[ $(artifact_name v1.0.0-rc.4) == dx-3fv-v1.0.0-rc.4.tar.gz ]]
[[ $(artifact_sha256 v1.0.0-rc.4) == 86493c9fac8500ef3e8d3504bc555ea14850e915f5f274012ed5f6f594406c17 ]]
[[ $(artifact_name v1.0.0) == dx-3fv-v1.0.0.tar.gz ]]
[[ $(artifact_sha256 v1.0.0) == 98e70f224329badf7315b415304ba08de24017f7304ecf2383a99d3921b41936 ]]
[[ $(artifact_name v1.0.1) == dx-3fv-v1.0.1.tar.gz ]]
[[ $(artifact_sha256 v1.0.1) == 207bba20d65b16614789aec0860f6b43d35089af829796852c44c743e3ad3f01 ]]
[[ $(artifact_name v1.0.2) == dx-3fv-v1.0.2.tar.gz ]]
[[ $(artifact_sha256 v1.0.2) == eb42ef55ed345f7c048dec1065790d4ee9ac28341648637bfb6590ed27822e78 ]]
[[ $(artifact_name v1.0.3) == dx-3fv-v1.0.3.tar.gz ]]
[[ $(artifact_sha256 v1.0.3) == 647143b02aed03d69272cf46cf6c7445c1e05c8d818c49b737293a643218ab69 ]]
[[ $(artifact_name v1.0.4) == dx-3fv-v1.0.4.tar.gz ]]
[[ $(artifact_sha256 v1.0.4) == 36d95737dc215404142df737c6cb25f269db72e8146082b497cd11f543cc5780 ]]
[[ $(artifact_name v1.0.5) == dx-3fv-v1.0.5.tar.gz ]]
[[ $(artifact_sha256 v1.0.5) == 94c80dc29efdeaecb3ed641b74fa6f8b08e73fcb1ece6b357440c530b20d7d2e ]]
[[ $(artifact_name v1.0.6) == dx-3fv-v1.0.6.tar.gz ]]
[[ $(artifact_sha256 v1.0.6) == 0ee969705e6c5096047247b865b27905e926d7e5abf744a4cf713407d9c5691d ]]
! artifact_name latest

archive="$repo_root/release/artifacts/$(artifact_name v1.0.0-rc.3)"
printf '%s  %s\n' "$(artifact_sha256 v1.0.0-rc.3)" "$archive" | sha256sum -c -
tar -tzf "$archive" 'dx-3fv-v1.0.0-rc.3/scripts/install-native.sh' >/dev/null

archive="$repo_root/release/artifacts/$(artifact_name v1.0.0-rc.4)"
printf '%s  %s\n' "$(artifact_sha256 v1.0.0-rc.4)" "$archive" | sha256sum -c -
tar -tzf "$archive" 'dx-3fv-v1.0.0-rc.4/scripts/install-native.sh' >/dev/null

archive="$repo_root/release/artifacts/$(artifact_name v1.0.0)"
printf '%s  %s\n' "$(artifact_sha256 v1.0.0)" "$archive" | sha256sum -c -
tar -tzf "$archive" 'dx-3fv-v1.0.0/scripts/install-native.sh' >/dev/null

archive="$repo_root/release/artifacts/$(artifact_name v1.0.1)"
printf '%s  %s\n' "$(artifact_sha256 v1.0.1)" "$archive" | sha256sum -c -
tar -tzf "$archive" 'dx-3fv-v1.0.1/scripts/install-native.sh' >/dev/null

for version in v1.0.2 v1.0.3 v1.0.4 v1.0.5 v1.0.6; do
  archive="$repo_root/release/artifacts/$(artifact_name "$version")"
  printf '%s  %s\n' "$(artifact_sha256 "$version")" "$archive" | sha256sum -c -
  tar -tzf "$archive" "dx-3fv-${version}/scripts/install-native.sh" >/dev/null
done

echo 'installer contract: PASS'
