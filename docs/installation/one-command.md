# Debian 12 一键安装

当前只支持 Debian 12 amd64。安装器下载固定 RC3 原生包并验证 SHA256；目标机只安装
Caddy、CA、curl、jq、OpenSSL、SQLite CLI、tar/gzip 等运行/运维依赖，不安装 Go、Node
或 pnpm。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/torr9522/dx-3fv/main/scripts/install.sh)
```

安装器会询问域名、ACME Staging/Production 和管理员凭据。测试安装请选择 TEST；测试
证书不受浏览器信任是正常现象。随机或自定义凭据仅在终端显示一次，并保存到 root-only
的 `/root/.config/dx-3fv/admin.env`（目录 0700、文件 0600）。

无人值守验收：

```bash
DX3FV_DOMAIN=panel.example.com \
DX3FV_ACME_MODE=staging \
DX3FV_ADMIN_MODE=random \
bash <(curl -fsSL https://raw.githubusercontent.com/torr9522/dx-3fv/main/scripts/install.sh) \
  --non-interactive --version v1.0.0
```

脚本检测到已有数据库、systemd 服务或 `/opt/dx-3fv/current` 时会退出，不会把重复安装
当作升级。升级和回滚继续使用发布包内的原生脚本。
