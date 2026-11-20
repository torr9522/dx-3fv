#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
version=${1:-$(git -C "$repo_root" describe --tags --always --dirty)}
output_root=${2:-"$repo_root/build/releases"}
output_root=$(realpath -m -- "$output_root")
release_name="dx-3fv-${version}"
stage="$output_root/$release_name"
go_toolchain=go1.26.11
schema_source="$repo_root/go-backend/internal/store/repo/repository.go"
schema_version=$(sed -nE 's/^const currentSchemaVersion = ([0-9]+)$/\1/p' "$schema_source")

if [[ ! "$schema_version" =~ ^[0-9]+$ ]]; then
  echo "unable to resolve currentSchemaVersion from $schema_source" >&2
  exit 2
fi

case "$version" in
  *[!A-Za-z0-9._+-]*) echo "invalid version: $version" >&2; exit 2 ;;
esac

if [[ -n "$(git -C "$repo_root" status --porcelain=v1)" ]]; then
  echo "refusing to build from a dirty worktree" >&2
  exit 2
fi
if [[ $(git -C "$repo_root" cat-file -t "refs/tags/$version" 2>/dev/null || true) != tag ]]; then
  echo "version must name an annotated tag: $version" >&2
  exit 2
fi
if [[ $(git -C "$repo_root" rev-list -n 1 "$version") != $(git -C "$repo_root" rev-parse HEAD) ]]; then
  echo "tag $version does not point to HEAD" >&2
  exit 2
fi

case "$output_root" in
  /|/root|/opt|/var|/usr|/etc) echo "unsafe output root: $output_root" >&2; exit 2 ;;
esac
[[ "$stage" == "$output_root"/dx-3fv-* ]] || { echo "unsafe stage: $stage" >&2; exit 2; }
if [[ -e "$stage" ]]; then
  rm -rf --one-file-system -- "$stage"
fi
install -d "$stage/bin" "$stage/frontend" "$stage/deploy" "$stage/scripts" \
  "$stage/sbom" "$stage/docs" "$stage/LICENSES" "$stage/migrations"

(
  cd "$repo_root/vite-frontend"
  VITE_APP_VERSION="$version" pnpm run build
)
(
  cd "$repo_root/go-backend"
  CGO_ENABLED=0 GOTOOLCHAIN="$go_toolchain" go build -trimpath -ldflags="-s -w" -o "$stage/bin/dx-3fv-paneld" ./cmd/paneld
)
(
  cd "$repo_root/go-gost"
  CGO_ENABLED=0 GOTOOLCHAIN="$go_toolchain" go build -trimpath -ldflags="-s -w -X main.version=$version" -o "$stage/bin/dx-3fv-agent" .
)

cp -a "$repo_root/vite-frontend/dist/." "$stage/frontend/"
cp "$repo_root/release/native/dx-3fv-panel.service" "$stage/deploy/"
cp "$repo_root/release/native/dx-3fv-agent.service" "$stage/deploy/"
cp "$repo_root/release/native/dx-3fv.caddy" "$stage/deploy/"
cp "$repo_root/release/native/panel.env.example" "$stage/deploy/"
cp "$repo_root/release/native/agent-config.json.example" "$stage/deploy/"
cp "$repo_root/release/native/agent-gost.json" "$stage/deploy/"
cp "$repo_root/release/native/install-native.sh" "$stage/scripts/"
cp "$repo_root/release/native/upgrade-native.sh" "$stage/scripts/"
cp "$repo_root/release/native/rollback-native.sh" "$stage/scripts/"
cp "$repo_root/release/native/health-check.sh" "$stage/scripts/"
cp "$repo_root/release/native/install-agent-native.sh" "$stage/scripts/"
cp "$repo_root/release/native/migrations/README.md" "$stage/migrations/"
cp "$repo_root/LICENSE" "$repo_root/LICENSE-APACHE" "$repo_root/NOTICE" \
  "$repo_root/README.md" "$repo_root/THIRD_PARTY_NOTICES.md" "$stage/"
cp -a "$repo_root/LICENSES/." "$stage/LICENSES/"
for docs_dir in architecture installation upgrade rollback operations security development release; do
  cp -a "$repo_root/docs/$docs_dir" "$stage/docs/"
done
chmod 0755 "$stage/bin/"* "$stage/scripts/"*.sh

commit=$(git -C "$repo_root" rev-parse HEAD)
tree=$(git -C "$repo_root" rev-parse HEAD^{tree})
cat >"$stage/MANIFEST" <<EOF
PRODUCT=DX-3FV
VERSION=$version
COMMIT=$commit
TREE=$tree
SCHEMA=$schema_version
GO=$(GOTOOLCHAIN="$go_toolchain" go version)
NODE=$(node --version)
PNPM=$(pnpm --version)
BUILT_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)
EOF

backend_modules=$(mktemp)
agent_modules=$(mktemp)
frontend_dependencies=$(mktemp)
frontend_licenses=$(mktemp)
cleanup_build_metadata() {
  rm -f -- "$backend_modules" "$agent_modules" \
    "$frontend_dependencies" "$frontend_licenses"
}
trap cleanup_build_metadata EXIT
(
  cd "$repo_root/go-backend"
  GOTOOLCHAIN="$go_toolchain" go list -m -json all >"$backend_modules"
)
(
  cd "$repo_root/go-gost"
  GOTOOLCHAIN="$go_toolchain" go list -m -json all >"$agent_modules"
)
node "$repo_root/scripts/sanitize-go-module-metadata.mjs" \
  "$backend_modules" "$stage/sbom/go-backend-modules.json"
node "$repo_root/scripts/sanitize-go-module-metadata.mjs" \
  "$agent_modules" "$stage/sbom/go-agent-modules.json"
(
  cd "$repo_root/vite-frontend"
  pnpm list --json --depth Infinity >"$frontend_dependencies"
  pnpm licenses list --json --prod >"$frontend_licenses"
)
node "$repo_root/scripts/sanitize-node-metadata.mjs" \
  "$frontend_dependencies" "$stage/sbom/frontend-dependencies.json"
node "$repo_root/scripts/sanitize-node-metadata.mjs" \
  "$frontend_licenses" "$stage/sbom/frontend-production-licenses.json"

(
  cd "$stage"
  find . -type f ! -name SHA256SUMS -print0 | sort -z | xargs -0 sha256sum >SHA256SUMS
)
tar -C "$output_root" -czf "$output_root/$release_name.tar.gz" "$release_name"
git -C "$repo_root" archive --format=tar.gz --prefix="$release_name-source/" \
  -o "$output_root/$release_name-source.tar.gz" "$version"
(
  cd "$output_root"
  sha256sum "$release_name.tar.gz" >"$release_name.tar.gz.sha256"
  sha256sum "$release_name-source.tar.gz" >"$release_name-source.tar.gz.sha256"
)

printf '%s\n' "$stage" "$output_root/$release_name.tar.gz" \
  "$output_root/$release_name-source.tar.gz"
