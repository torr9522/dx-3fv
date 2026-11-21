# 发布历史

功能 Tag 和问题映射以 `git tag --format='%(refname:short) %(contents:subject)'` 与
`docs/issues/` 为准。冻结候选：

- `v0.1.48-fix-frontend-runtime-dependencies`：业务冻结基线，Schema 12。
- `v1.0.0-rc.1`：构建成功但 SBOM 含构建机路径，淘汰。
- `v1.0.0-rc.2`：修复路径清洗但正式构建在 Go template 阶段失败，淘汰。
- `v1.0.0-rc.3`：修复 JSON stream 解析，通过原生构建、安装/升级/回滚和真实业务验收。
- `v1.0.0-rc.4`：Phase 5 功能收尾、Router 7、小时流量、100 并发和故障恢复候选。

Phase 4/5 的 publication Tag 只表示公开历史或安装器候选，不得冒充最终稳定版。
`v1.0.0` 以 RC4 已验证源码为基线，仅补充正式发布材料和固定安装器映射，不增加产品
功能或改变 Schema、Agent 协议及 Runtime。

Phase 5 独立交付：

- `v0.1.60-fix-multiport-create-scope`：新增业务入口与 PathProfile 作用域。
- `v0.1.61-fix-udp-session-expiry`：修复 UDP session 到期/入队竞态。
- `v0.1.62-migrate-react-router-7`：Declarative Mode 迁移到 React Router 7.18.6。
- `v0.1.63-verify-hour-scale-traffic`：61 分钟真实持续流量证据。
- `v0.1.64-fix-full-pool-allocation-scan`：100 并发验收与 full Pool 自动分配阻塞修复。
- `v0.1.65-verify-fault-recovery-matrix`：完整可恢复故障注入矩阵及最终业务回归证据。
- `v1.0.0`：首个稳定版；保留公网 UDP 网络依赖限制，不承诺零丢包或无限容量。
