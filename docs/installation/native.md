# 原生安装

## 前提

- 64 位 Linux 与 systemd；RC 在 Debian 12 amd64 验证。
- root 权限、`sqlite3`、`openssl`、`curl`、`ca-certificates`、Caddy。
- 一个指向 Panel 的域名及可签发证书的 80/443 入站路径。
- 同版本二进制包、源码包及 SHA256 文件。

不要在目标机安装 Docker/Podman。不要复用其他 FLVX 实例的数据库、JWT Secret 或
Agent Secret。

## 校验与解压

```bash
sha256sum -c dx-3fv-VERSION.tar.gz.sha256
sha256sum -c dx-3fv-VERSION-source.tar.gz.sha256
tar -xzf dx-3fv-VERSION.tar.gz
cd dx-3fv-VERSION
sha256sum -c SHA256SUMS
```

检查 `MANIFEST` 中版本、Commit、Tree 和 Schema 与 Release Notes 一致。

## 安装 Panel

```bash
sudo ./scripts/install-native.sh "$PWD"
```

脚本创建 `dx-3fv` 系统用户、`/opt/dx-3fv/releases`、`/var/lib/dx-3fv`、
`/var/log/dx-3fv` 和 `/etc/dx-3fv/panel.env`，生成 JWT Secret，安装 systemd unit，
并执行本机健康检查。配置文件权限必须保持 `0600`。

将 `deploy/dx-3fv.caddy` 复制到 Caddy 配置目录前，把 `panel.example.com` 替换为
真实域名。Caddy 应把 `/api/*`、`/flow/*`、`/error`、`/system-info` 反代到
`127.0.0.1:17365`，其他请求从 `/opt/dx-3fv/current/frontend` 提供。

通行证密钥默认关闭。需要启用时，在 `/etc/dx-3fv/panel.env` 设置：

```dotenv
FLVX_WEBAUTHN_ORIGIN=https://your-panel.example
```

该值不能含路径、查询参数或尾部斜杠。重启 Panel 后生效。

## 安装 Agent

先在 Panel 创建唯一 Node，记录 Node ID、显示名称和节点专属 Secret。基于
`deploy/agent-config.json.example` 在节点本机创建 `config.json`，Panel 地址使用
可校验证书的 HTTPS 地址，文件权限设为 `0600`。

```bash
sudo ./scripts/install-agent-native.sh "$PWD" /secure/path/config.json
```

脚本要求非空 CA bundle，安装空 GOST Runtime 配置，启动 Agent，并等待连续三次
systemd running 状态。不要把一个节点的配置复制给另一个节点。

## 验收

```bash
systemctl status dx-3fv-panel.service --no-pager
./scripts/health-check.sh
systemctl status dx-3fv-agent.service --no-pager
```

随后从浏览器使用 HTTPS 登录，立即修改初始密码，确认节点 ONLINE，再创建最小测试
业务并核对 Listener、TCP/UDP 能力和流量。Agent ONLINE 不是业务验收的替代品。
