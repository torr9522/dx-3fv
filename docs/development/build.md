# 可复现原生构建

## 工具链

- Go `1.26.11`，由 `GOTOOLCHAIN` 固定。
- Node `.node-version` 指定 `22.23.4`；也允许 `24.22.1+` 或 `26.12.1+` 的已修复版本。
- pnpm `10.28.1`，依赖由 `pnpm-lock.yaml` 锁定。

## 前置检查

工作区必须 clean，版本参数必须是位于 HEAD 的 Annotated Tag。先运行：

```bash
GOTOOLCHAIN=go1.26.11 go test ./...   # 分别在 go-backend、go-gost、go-gost/x
cd vite-frontend && pnpm install --frozen-lockfile && pnpm run build
```

## 构建

```bash
./scripts/build-native-release.sh v1.0.0-rc.1 /absolute/output
```

脚本构建静态 Panel/Agent、前端，读取源码中的 Schema，生成 Manifest、Go module
清单、脱敏 pnpm dependency tree、生产依赖许可证清单、逐文件 SHA256、二进制归档
和同 Commit 源码归档。

## 验证

解压到临时目录运行 `sha256sum -c SHA256SUMS`，确认没有绝对构建路径、真实域名、
凭据或数据库。比较源码归档 HEAD 与 Manifest Commit，并在隔离目录完成安装、升级、
回滚演练。构建成功本身不是发布验收。
