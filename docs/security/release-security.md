# 发布安全

## Secret 边界

发布包和 Git 不得包含 SSH 凭据、API Key、JWT Secret、Agent Secret、SS URI、
私钥、真实数据库、环境文件、日志或 known_hosts 证据。只发布占位模板。最终发布前
同时扫描工作树、全部 Git objects 和二进制/源码归档。

`panel.env`、Agent `config.json` 与备份必须为 `0600` 或更严格。不要在工单、截图、
命令输出或健康检查中打印它们。

## 网络边界

- Panel 公开入口使用 HTTPS；后端默认只监听 `127.0.0.1`。
- Agent 使用可验证的 CA bundle，不允许为方便关闭证书验证。
- 管理连接和业务数据面分离。不要把节点出口公网地址当作 SSH 备用入口。
- 诊断只访问当前业务持久化的精确目标，不扩展成扫描。

## 供应链

构建只接受 clean worktree 上指向 HEAD 的 Annotated Tag。Go 工具链固定为
`1.26.11`；Node/pnpm 版本和 Commit/Tree 写入 Manifest。每个发布文件进入
`SHA256SUMS`，外层归档另有 SHA256。SBOM 中构建机绝对路径会被移除。

依赖公告按可达性处理，不能为清零 audit 数字盲目升级主版本。当前延期项见 Release
Notes 和 Known Issues。
