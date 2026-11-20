# DX3FV-0040：全局默认 IPv4-only

- 基线：`v0.1.43-fix-diagnosis-echo-semantics`
- 数据库 Schema：不变（Schema 12）
- 系统网络栈：不修改 sysctl、防火墙或接口

## 调查结论

旧产品默认把新节点的 TCP/UDP Listener 保存为 `[::]`，Tunnel 的
`ipPreference=v4` 在无法找到 IPv4 共同地址时还会回退 IPv6。普通
Forward 和多端口业务的后端也接受 IPv6 Landing，因此只隐藏前端字段
无法形成可靠的 IPv4-only 策略。

GOST Fork 内部仍有完整的 IPv6 Listener、Dialer 和协议处理能力。直接
删除这些能力会破坏历史配置及底层通用 Runtime，所以本 Issue 只约束
DX-3FV 新建和修改业务的产品边界。

## 修复边界

- 新节点 TCP/UDP Listener 默认 `0.0.0.0`。
- 新节点拒绝 IPv6 地址；域名必须解析到至少一个 IPv4 A 记录。
- 新 Tunnel 固定保存 `ipPreference=v4`，并且不再回退 IPv6。
- 新 Tunnel 的自定义连接地址、入口地址和诊断目标必须可使用 IPv4。
- 新普通 Forward 和多端口 Landing 拒绝 IPv6；域名由后端确认 A 记录。
- 前端移除新建 IPv6 操作入口，Landing 和诊断表单明确显示 IPv4-only。
- 历史 IPv6 Node/Tunnel/Forward 仍可读取；编辑时未改变的历史值可原样
  保留，避免升级时重置正在运行的旧服务。
- Federation 新导入节点及空 Listener 的控制面回退值改为 `0.0.0.0`。

## 代码路径

- `go-backend/internal/http/handler/ipv4_policy.go`：统一地址、A 记录、
  Listener、Forward 和 Tunnel 校验。
- `go-backend/internal/http/handler/mutations.go`：Node、Tunnel、Forward
  新建/编辑边界及历史原值兼容。
- `go-backend/internal/http/handler/ix_routes.go`：多端口 Landing 与 Listener。
- `go-backend/internal/store/model/model.go`、`repository_control.go`、
  `repository_federation.go`：新记录和空值默认。
- `vite-frontend/src/pages/node.tsx`、`tunnel.tsx`、`multiport.tsx`、
  `pages/forward/landing-target.ts`：用户操作层。

## 自动化验收

- IPv4 literal、A-only domain、AAAA-only domain、IPv6 literal 单元测试。
- `ipPreference` 契约覆盖：空值默认 v4、新建 v6 被拒、历史 v6 列表可读。
- Handler 全套测试。
- Frontend TypeScript + production build。
- Landing 表单保留历史 IPv6 解析，但新建校验拒绝 IPv6。

## 兼容与回滚

回滚到 `v0.1.43-fix-diagnosis-echo-semantics`。本 Issue 不迁移或删除任何
历史 IPv6 数据，也不修改操作系统网络配置；回滚只恢复新建默认值和
操作入口。
