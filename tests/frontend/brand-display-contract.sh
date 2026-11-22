#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
frontend="$repo_root/vite-frontend"

require_text() {
  local file=$1
  local text=$2
  grep -Fq -- "$text" "$file"
}

reject_text() {
  local file=$1
  local text=$2
  ! grep -Fq -- "$text" "$file"
}

require_text "$frontend/index.html" '<title>DX-3FV</title>'
require_text "$frontend/index.html" 'apple-mobile-web-app-title" content="DX-3FV"'
reject_text "$frontend/index.html" 'content="FLVX"'
require_text "$frontend/vite.config.ts" 'name: "DX-3FV"'
require_text "$frontend/vite.config.ts" 'short_name: "DX-3FV"'
reject_text "$frontend/vite.config.ts" '"FLVX"'
require_text "$frontend/src/config/site.ts" 'const DEFAULT_BRAND_NAME = "DX-3FV";'
reject_text "$frontend/src/config/site.ts" 'name: "FLVX"'
require_text "$frontend/src/components/navbar.tsx" 'await updateSiteConfig();'
reject_text "$frontend/src/components/navbar.tsx" 'getCachedConfig("app_name")'
require_text "$frontend/src/components/version-footer.tsx" 'DX-3FV'
reject_text "$frontend/src/components/version-footer.tsx" 'FLVX'
require_text "$frontend/src/pages/config.tsx" '隐藏页面底部 DX-3FV 品牌信息'
require_text "$frontend/src/pages/forward.tsx" '兼容格式（管道分隔）'
reject_text "$frontend/src/pages/forward.tsx" 'flvx格式'
require_text "$frontend/src/themes/default/index.ts" 'author: "DX-3FV Team"'
require_text "$frontend/src/themes/example-cyberpunk/index.ts" 'author: "DX-3FV Community"'

printf 'frontend brand display contract: PASS\n'
