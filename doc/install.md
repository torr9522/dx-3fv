# 安装部署指南（已迁移）

DX-3FV 正式版仅支持经过校验的 Debian 12 amd64 原生部署，不使用继承的 FLVX
Docker 安装器或上游 Release 作为更新源。

正式一键安装命令：

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/torr9522/dx-3fv/main/scripts/install.sh)
```

安装器默认选择固定的正式版本，不跟随未发布的 `main` 构建。完整的安装、Agent
部署、升级和回滚步骤分别见：

- `docs/installation/one-command.md`
- `docs/installation/native-deployment.md`
- `docs/upgrade/native.md`
- `docs/rollback/native.md`

根目录 `panel_install.sh`、`install.sh` 与 Docker Compose 模板仅作为明确停用的兼容
占位文件，执行时不会下载或安装上游 FLVX 资产。
