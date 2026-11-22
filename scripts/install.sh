#!/usr/bin/env bash
set -Eeuo pipefail

INSTALLER_VERSION="1.0.4"
DEFAULT_VERSION="v1.0.4"
REPOSITORY="torr9522/dx-3fv"
PANEL_ADDR="127.0.0.1:17365"
STAGING_CA="https://acme-staging-v02.api.letsencrypt.org/directory"
WORK_DIR=""
INSTALL_STARTED=false
INSTALL_COMPLETE=false
CADDY_SITE_CREATED=false
CADDY_IMPORT_ADDED=false

log() { printf '[dx-3fv] %s\n' "$*"; }
fail() { printf '[dx-3fv] ERROR: %s\n' "$*" >&2; exit 1; }

remove_work_dir() {
  if [[ -n $WORK_DIR ]]; then
    case "$WORK_DIR" in
      /var/tmp/dx-3fv-install.*) rm -rf -- "$WORK_DIR" ;;
      *) log "拒绝清理非安装临时目录: $WORK_DIR" ;;
    esac
  fi
}

rollback_failed_install() {
  local exit_code=$?
  if [[ $exit_code -ne 0 && $INSTALL_STARTED == true && $INSTALL_COMPLETE == false ]]; then
    log "安装失败，回滚本次新建的 DX-3FV 文件"
    systemctl disable --now dx-3fv-panel.service >/dev/null 2>&1 || true
    if [[ $CADDY_SITE_CREATED == true ]]; then
      rm -f -- /etc/caddy/conf.d/dx-3fv.caddy
    fi
    if [[ $CADDY_IMPORT_ADDED == true && -f /etc/caddy/Caddyfile.dx3fv-preinstall ]]; then
      cp -a /etc/caddy/Caddyfile.dx3fv-preinstall /etc/caddy/Caddyfile
    fi
    systemctl restart caddy >/dev/null 2>&1 || true
    rm -f -- /etc/systemd/system/dx-3fv-panel.service /opt/dx-3fv/current
    rm -rf -- "/opt/dx-3fv/releases/dx-3fv-${SELECTED_VERSION}" /etc/dx-3fv /var/lib/dx-3fv /var/log/dx-3fv
    systemctl daemon-reload >/dev/null 2>&1 || true
  fi
  remove_work_dir
  return "$exit_code"
}

trap rollback_failed_install EXIT

usage() {
  cat <<'EOF'
DX-3FV one-command installer (Debian 12 amd64)

Usage: bash install.sh [--version VERSION] [--ref GIT_REF] [--non-interactive]

Environment for unattended acceptance:
  DX3FV_DOMAIN          Required domain
  DX3FV_ACME_MODE       staging or production
  DX3FV_ADMIN_MODE      random or custom
  DX3FV_ADMIN_USERNAME  Required with custom mode
  DX3FV_ADMIN_PASSWORD  Required with custom mode

The stable installer defaults to the fixed v1.0.4 tag; --version never
follows latest or an unversioned development artifact.
EOF
}

validate_domain() {
  local value=${1,,}
  [[ ${#value} -le 253 ]] || return 1
  [[ $value =~ ^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$ ]]
}

validate_username() {
  [[ $1 =~ ^[A-Za-z0-9_.-]{4,64}$ ]]
}

validate_password() {
  local value=$1
  [[ ${#value} -ge 16 && ${#value} -le 128 ]] || return 1
  [[ $value =~ [a-z] ]] || return 1
  [[ $value =~ [A-Z] ]] || return 1
  [[ $value =~ [0-9] ]] || return 1
  [[ $value =~ [^A-Za-z0-9] ]]
}

artifact_name() {
  case "$1" in
    v1.0.0-rc.3) printf 'dx-3fv-v1.0.0-rc.3.tar.gz\n' ;;
    v1.0.0-rc.4) printf 'dx-3fv-v1.0.0-rc.4.tar.gz\n' ;;
    v1.0.0) printf 'dx-3fv-v1.0.0.tar.gz\n' ;;
    v1.0.1) printf 'dx-3fv-v1.0.1.tar.gz\n' ;;
    v1.0.2) printf 'dx-3fv-v1.0.2.tar.gz\n' ;;
    v1.0.3) printf 'dx-3fv-v1.0.3.tar.gz\n' ;;
    v1.0.4) printf 'dx-3fv-v1.0.4.tar.gz\n' ;;
    *) return 1 ;;
  esac
}

artifact_sha256() {
  case "$1" in
    v1.0.0-rc.3) printf '85d5aa4e5e68b2d286bfae08efc97a0a3e16febaa2684ff8226c95c43b59874f\n' ;;
    v1.0.0-rc.4) printf '86493c9fac8500ef3e8d3504bc555ea14850e915f5f274012ed5f6f594406c17\n' ;;
    v1.0.0) printf '98e70f224329badf7315b415304ba08de24017f7304ecf2383a99d3921b41936\n' ;;
    v1.0.1) printf '207bba20d65b16614789aec0860f6b43d35089af829796852c44c743e3ad3f01\n' ;;
    v1.0.2) printf 'eb42ef55ed345f7c048dec1065790d4ee9ac28341648637bfb6590ed27822e78\n' ;;
    v1.0.3) printf '647143b02aed03d69272cf46cf6c7445c1e05c8d818c49b737293a643218ab69\n' ;;
    v1.0.4) printf '36d95737dc215404142df737c6cb25f269db72e8146082b497cd11f543cc5780\n' ;;
    *) return 1 ;;
  esac
}

read_domain() {
  local value=${DX3FV_DOMAIN:-}
  while [[ -z $value ]]; do
    read -r -p '请输入 DX-3FV 访问域名: ' value
  done
  value=${value,,}
  validate_domain "$value" || fail "域名格式无效: $value"
  printf '%s\n' "$value"
}

read_acme_mode() {
  local value=${DX3FV_ACME_MODE:-}
  if [[ -z $value ]]; then
    printf '证书模式: 1) TEST / ACME Staging  2) PRODUCTION\n' >&2
    read -r -p '请选择 [1]: ' value
    value=${value:-1}
  fi
  case "${value,,}" in
    1|test|staging) printf 'staging\n' ;;
    2|production|prod) printf 'production\n' ;;
    *) fail "证书模式必须是 staging 或 production" ;;
  esac
}

random_admin_username() {
  printf 'dxadmin_%s\n' "$(openssl rand -hex 5)"
}

random_admin_password() {
  printf 'D3!%sAa9#\n' "$(openssl rand -hex 12)"
}

read_admin_credentials() {
  local mode=${DX3FV_ADMIN_MODE:-}
  if [[ -z $mode ]]; then
    read -r -p '是否自定义管理员用户名和密码? [y/N]: ' mode
    case "${mode,,}" in y|yes) mode=custom ;; *) mode=random ;; esac
  fi

  case "${mode,,}" in
    random)
      ADMIN_USERNAME=$(random_admin_username)
      ADMIN_PASSWORD=$(random_admin_password)
      ;;
    custom)
      ADMIN_USERNAME=${DX3FV_ADMIN_USERNAME:-}
      ADMIN_PASSWORD=${DX3FV_ADMIN_PASSWORD:-}
      if [[ -z $ADMIN_USERNAME ]]; then
        read -r -p '管理员用户名: ' ADMIN_USERNAME
      fi
      if [[ -z $ADMIN_PASSWORD ]]; then
        read -r -s -p '管理员密码: ' ADMIN_PASSWORD
        printf '\n' >&2
        local confirm
        read -r -s -p '再次输入管理员密码: ' confirm
        printf '\n' >&2
        [[ $ADMIN_PASSWORD == "$confirm" ]] || fail "两次密码不一致"
      fi
      ;;
    *) fail "DX3FV_ADMIN_MODE 必须是 random 或 custom" ;;
  esac

  validate_username "$ADMIN_USERNAME" || fail "管理员用户名需为 4-64 位字母、数字、点、下划线或连字符"
  validate_password "$ADMIN_PASSWORD" || fail "管理员密码需 16-128 位并包含大小写字母、数字和符号"
}

require_clean_target() {
  [[ ! -e /opt/dx-3fv ]] || fail "检测到现有 /opt/dx-3fv；请使用升级流程"
  [[ ! -e /etc/dx-3fv ]] || fail "检测到现有 /etc/dx-3fv；不会覆盖"
  [[ ! -e /var/lib/dx-3fv ]] || fail "检测到现有 /var/lib/dx-3fv；不会覆盖"
  [[ ! -e /var/log/dx-3fv ]] || fail "检测到现有 /var/log/dx-3fv；不会覆盖"
  [[ ! -e /etc/systemd/system/dx-3fv-panel.service ]] || fail "检测到现有 systemd 服务；不会覆盖"
  if systemctl list-unit-files dx-3fv-panel.service --no-legend 2>/dev/null | grep -q dx-3fv; then
    fail "检测到现有 DX-3FV service；不会覆盖"
  fi
  if ss -H -ltn "sport = :17365" 2>/dev/null | grep -q .; then
    fail "Panel 本地端口 17365 已被占用"
  fi
}

preflight() {
  [[ $EUID -eq 0 ]] || fail "必须以 root 运行"
  [[ $(uname -m) == x86_64 ]] || fail "当前只支持 amd64/x86_64"
  [[ -r /etc/os-release ]] || fail "无法识别操作系统"
  # shellcheck disable=SC1091
  . /etc/os-release
  [[ ${ID:-} == debian && ${VERSION_ID:-} == 12 ]] || fail "当前只支持 Debian 12"
  [[ $(df -Pk /opt 2>/dev/null | awk 'NR==2 {print $4}') -ge 524288 ]] || fail "可用磁盘空间不足 512 MiB"
  [[ $(awk '/MemTotal:/ {print $2}' /proc/meminfo) -ge 524288 ]] || fail "内存不足 512 MiB"
  require_clean_target

  if ss -H -ltn '( sport = :80 or sport = :443 )' 2>/dev/null | grep -q .; then
    systemctl is-active --quiet caddy || fail "80/443 已被非 Caddy 服务占用"
  fi
}

install_dependencies() {
  export DEBIAN_FRONTEND=noninteractive
  apt-get update
  apt-get install -y --no-install-recommends ca-certificates curl caddy jq openssl sqlite3 tar gzip iproute2
}

verify_dns() {
  local domain=$1 current_ip dns_ips
  current_ip=$(curl -4 -fsS --proto '=https' --tlsv1.2 --max-time 15 https://api.ipify.org)
  [[ $current_ip =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]] || fail "无法确认当前服务器公网 IPv4"
  dns_ips=$(getent ahostsv4 "$domain" | awk '{print $1}' | sort -u)
  [[ -n $dns_ips ]] || fail "域名没有可用的 IPv4 A 记录"
  grep -Fxq "$current_ip" <<<"$dns_ips" || fail "域名 A 记录未指向当前服务器 IPv4 $current_ip"
}

download_release() {
  local version=$1 ref=$2 output=$3 name expected url
  name=$(artifact_name "$version") || fail "不支持的固定版本: $version"
  expected=$(artifact_sha256 "$version") || fail "版本缺少 SHA256: $version"
  url="https://raw.githubusercontent.com/${REPOSITORY}/${ref}/release/artifacts/${name}"
  log "下载并校验 ${version}"
  curl -fL --retry 3 --connect-timeout 15 --proto '=https' --tlsv1.2 "$url" -o "$output"
  printf '%s  %s\n' "$expected" "$output" | sha256sum -c -
}

initialize_admin() {
  local login_file token code update_file
  [[ -n $WORK_DIR && -d $WORK_DIR ]] || fail "管理员初始化临时目录不存在"
  login_file="$WORK_DIR/admin-login.json"
  update_file="$WORK_DIR/admin-update.json"
  : >"$login_file"
  : >"$update_file"
  chmod 0600 "$login_file" "$update_file"

  curl -fsS -H 'Content-Type: application/json' \
    --data '{"username":"admin_user","password":"admin_user"}' \
    "http://${PANEL_ADDR}/api/v1/user/login" >"$login_file"
  code=$(jq -r '.code // -1' "$login_file")
  [[ $code == 0 ]] || fail "默认管理员登录失败，拒绝继续"
  token=$(jq -r '.data.token // empty' "$login_file")
  [[ -n $token ]] || fail "管理员初始化未返回 token"

  jq -n --arg username "$ADMIN_USERNAME" --arg password "$ADMIN_PASSWORD" \
    '{newUsername:$username,currentPassword:"admin_user",newPassword:$password,confirmPassword:$password}' |
    curl -fsS -H 'Content-Type: application/json' -H "Authorization: $token" \
      --data-binary @- "http://${PANEL_ADDR}/api/v1/user/updatePassword" >"$update_file"
  code=$(jq -r '.code // -1' "$update_file")
  [[ $code == 0 ]] || fail "管理员账号初始化失败"

  : >"$login_file"
  : >"$update_file"
}

save_admin_credentials() {
  install -d -m 0700 /root/.config/dx-3fv
  umask 077
  {
    printf 'DX3FV_ADMIN_USERNAME=%q\n' "$ADMIN_USERNAME"
    printf 'DX3FV_ADMIN_PASSWORD=%q\n' "$ADMIN_PASSWORD"
  } >/root/.config/dx-3fv/admin.env
  chmod 0600 /root/.config/dx-3fv/admin.env
}

configure_caddy() {
  local domain=$1 mode=$2 caddyfile=/etc/caddy/Caddyfile site=/etc/caddy/conf.d/dx-3fv.caddy
  install -d -m 0755 /etc/caddy/conf.d
  if grep -R -F --include='*.caddy' --include='Caddyfile' "$domain" /etc/caddy 2>/dev/null | grep -q .; then
    fail "Caddy 已存在该域名配置，拒绝覆盖"
  fi
  if ! grep -Fq 'import /etc/caddy/conf.d/*.caddy' "$caddyfile"; then
    [[ ! -e ${caddyfile}.dx3fv-preinstall ]] || fail "存在旧 Caddy 安装备份，请先人工核对"
    cp -a "$caddyfile" "${caddyfile}.dx3fv-preinstall"
    printf '\nimport /etc/caddy/conf.d/*.caddy\n' >>"$caddyfile"
    CADDY_IMPORT_ADDED=true
  fi

  {
    printf '%s {\n' "$domain"
    if [[ $mode == staging ]]; then
      printf '    tls {\n        ca %s\n    }\n' "$STAGING_CA"
    fi
    cat <<'EOF'
    encode zstd gzip
    @backend path /api/* /flow/* /error /system-info
    handle @backend {
        reverse_proxy 127.0.0.1:17365
    }
    handle {
        root * /opt/dx-3fv/current/frontend
        try_files {path} /index.html
        file_server
    }
    header {
        X-Content-Type-Options nosniff
        Referrer-Policy strict-origin-when-cross-origin
    }
}
EOF
  } >"$site"
  CADDY_SITE_CREATED=true
  chmod 0644 "$site"
  caddy validate --config "$caddyfile"
  systemctl enable caddy
  systemctl restart caddy
}

wait_for_https() {
  local domain=$1 mode=$2 curl_args=(-fsS --max-time 5)
  [[ $mode == staging ]] && curl_args+=(-k)
  for _ in $(seq 1 36); do
    if curl "${curl_args[@]}" "https://${domain}/" >/dev/null 2>&1; then
      return 0
    fi
    sleep 5
  done
  systemctl status caddy --no-pager -l >&2 || true
  fail "HTTPS 在 180 秒内未就绪"
}

main() {
  local version=$DEFAULT_VERSION ref=main non_interactive=false
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --version) [[ $# -ge 2 ]] || fail "--version 缺少值"; version=$2; shift 2 ;;
      --ref) [[ $# -ge 2 ]] || fail "--ref 缺少值"; ref=$2; shift 2 ;;
      --non-interactive) non_interactive=true; shift ;;
      --help|-h) usage; exit 0 ;;
      *) fail "未知参数: $1" ;;
    esac
  done
  if $non_interactive; then
    [[ -n ${DX3FV_DOMAIN:-} && -n ${DX3FV_ACME_MODE:-} && -n ${DX3FV_ADMIN_MODE:-} ]] ||
      fail "非交互模式缺少 DX3FV_DOMAIN/DX3FV_ACME_MODE/DX3FV_ADMIN_MODE"
  fi

  artifact_name "$version" >/dev/null || fail "不支持的固定版本: $version"
  [[ $ref =~ ^[A-Za-z0-9._/-]+$ && $ref != *..* ]] || fail "无效 Git ref"
  SELECTED_VERSION=$version
  preflight
  local domain acme_mode work_dir archive release_dir
  domain=$(read_domain)
  acme_mode=$(read_acme_mode)
  install_dependencies
  read_admin_credentials
  verify_dns "$domain"

  work_dir=$(mktemp -d /var/tmp/dx-3fv-install.XXXXXX)
  WORK_DIR=$work_dir
  archive="$work_dir/$(artifact_name "$version")"
  download_release "$version" "$ref" "$archive"
  tar -xzf "$archive" -C "$work_dir"
  release_dir="$work_dir/dx-3fv-${version}"
  [[ -d $release_dir ]] || fail "发布包目录结构无效"
  INSTALL_STARTED=true
  "$release_dir/scripts/install-native.sh" "$release_dir"
  initialize_admin
  save_admin_credentials
  configure_caddy "$domain" "$acme_mode"
  wait_for_https "$domain" "$acme_mode"
  find /var/lib/dx-3fv -maxdepth 1 -type f -name 'dx-3fv.db*' -exec chmod 0600 {} +
  INSTALL_COMPLETE=true
  remove_work_dir
  WORK_DIR=""

  log "安装成功"
  [[ $acme_mode == staging ]] && log "当前使用 ACME Staging 测试证书，浏览器不信任是正常现象"
  printf '访问地址: https://%s\n' "$domain"
  printf '管理员用户名: %s\n' "$ADMIN_USERNAME"
  printf '管理员密码: %s\n' "$ADMIN_PASSWORD"
  printf '凭据保存位置: /root/.config/dx-3fv/admin.env\n'
  printf '安装器版本: %s\n' "$INSTALLER_VERSION"
}

if [[ ${BASH_SOURCE[0]} == "$0" ]]; then
  main "$@"
fi
