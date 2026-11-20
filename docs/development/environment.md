# 开发环境

支持 Linux 原生开发。Go 三个 module 使用 `1.26.11`；Node 由 `.node-version` 固定为
`22.23.4`；pnpm 为 `10.28.1`。前端依赖以 `pnpm-lock.yaml` 为唯一锁定来源。

数据库默认 SQLite，PostgreSQL 契约需要显式测试 DSN。测试凭据放在 Git 外部且不得
写入 shell history、日志或 fixture。不要用容器替代正式原生构建和安装验收。
