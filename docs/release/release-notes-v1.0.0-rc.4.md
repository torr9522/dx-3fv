# DX-3FV v1.0.0-rc.4 Release Notes

## 定位

`v1.0.0-rc.4` 是 Phase 5 功能收尾后的发布候选。数据库仍为 Schema 12，原 Forward、
Tunnel、多端口模型、Agent 协议和 GOST 架构保持兼容。本候选不创建正式 `v1.0.0`，
等待用户在测试站体验确认。

## 本候选完成项

- 多端口页面增加全局“新增业务”，分组内创建由前后端共同约束 PathProfile；保留独立
  “新增路径”入口和已有数据。
- 修复 UDP session 到期/关闭与 datagram 入队竞态，防止已失效 session 接收新 payload。
- Declarative Mode 从 React Router 6.30.6 迁移到固定的 7.18.6，不引入 SSR 或
  Framework Mode。
- 完成 3666 秒真实持续流量：10697 个 payload，UDP 8738/8745、TCP 1940/1952，
  没有节点离线、进程重启或 SQLite 错误。
- 完成七场景最高 100 并发验证；修复 full Pool 自动分配穷举约 65K Resource、阻塞
  SQLite 单连接和全部 Panel 请求的问题。
- 完成 Panel、SQLite、四 Agent、内嵌 GOST、部分创建/删除、Lease、端口冲突、编辑
  回滚、QIX 共享 Relay 和失败 Agent 升级的真实故障恢复矩阵。

## 发布验证

- 后端全套、定向并发测试、Agent lifecycle、前端 TypeScript/build 均通过。
- 真实测试 Panel 最终四节点 ONLINE，7 条 active Route、20 条 applied Runtime、
  20 个 active Lease，SQLite integrity `ok`，Schema 12。
- 最终七场景 payload 覆盖单端口、前海 TCP/UDP/混合、QIX 共享 Relay、原版 Tunnel
  和多个 PathProfile，7/7 通过。
- 当前源码 Tree 与完整 sanitized publication history 的 Gitleaks 扫描为 0 findings。

## 已知限制

- 61 分钟窗口记录 7 次 UDP 单包超时；均自行恢复，但本候选不提供零丢包 SLA。
- 100 并发窗口记录 4 次外部 Cloudflare TLS handshake failure；最大验证级别不是容量
  上限，也不代表无限并发。
- 精确 goroutine 值未暴露；资源结论限于实际采样窗口。
- 主要发布平台仍为 Debian 12 amd64、systemd、Caddy、SQLite；其他发行版、架构、
  PostgreSQL 大数据量和多 Panel 副本需另行验收。
- 原开发仓库含三个归属不明的已删除上游配置候选值，因此不公开其 ancestry；公开仓库
  使用有明确血缘说明的 sanitized snapshot history。

详情见 `known-issues.md`、DX3FV-0057 至 DX3FV-0062 Issue 文档以及安全/许可证文档。

## 发布边界

RC4 原生包包含 Panel、Agent、前端、systemd、配置模板、安装/升级/回滚/健康检查、
Manifest、逐文件 SHA256、SBOM、LICENSE、NOTICE 和中文运维文档。安装器只接受显式
版本及内置 SHA256，不跟随 GitHub latest。禁止容器部署。
