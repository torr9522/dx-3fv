# DX3FV-0055：Debian 12 一键安装器

- 类型：Release installation
- 状态：实现完成，待纯净服务器验收
- 产品/Schema/Agent/GOST 影响：无

## 设计依据

RC3 已有可校验原生包和低层 `install-native.sh`，但缺少依赖、域名、TLS、Caddy 和管理员
初始化编排。新脚本复用 RC3 原生安装，不在目标机安装开发工具链或重新实现 Panel 部署。

## 实现

- Debian 12 amd64/root/磁盘/内存/端口/已有安装检查。
- 固定版本映射、raw GitHub 下载和内置 SHA256 fail-closed 校验。
- DNS A 与当前服务器 IPv4 一致性检查。
- ACME Staging/Production，独立 Caddy 站点，Panel 仅监听 `127.0.0.1`。
- CSPRNG 随机管理员或强度校验的自定义凭据；通过本地 API 立即替换默认账号。
- root-only 凭据文件、数据库权限、HTTPS 和 systemd health check。

## 自动化测试

- `bash -n scripts/install.sh`。
- `tests/install/installer-contract.sh`：输入校验、固定版本、SHA256、归档结构。
- publication gitleaks 全历史扫描。

## 真人验收

在 `192.255.175.132` 重装 Debian 12 后，仅从 GitHub 执行安装命令；测试域名使用 ACME
Staging。此项完成后补充次数、Host Key、登录、重启恢复和浏览器证据。

## 回滚

未安装时回退本提交。已安装主机按 `docs/rollback/native.md` 恢复；安装器不会自动删除
数据库或覆盖既有服务。
