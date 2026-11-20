# 架构决策记录

1. 保留普通 Forward：亚太入口直连 Landing 不需要 Tunnel 或 IX。
2. 多端口是独立域：PathProfile 组合节点，IXRoute 拥有 Lease/Runtime，不伪造 Tunnel。
3. 端口以 Resource/Lease/Owner 管理：Provider 公网映射与内部 Listener 分离。
4. 共享 Relay 复用真实 Type 2 Tunnel：每个 Forward 独立 Service，删除隔离。
5. 数据库成功不等于 Runtime 成功：Desired/Observed、Observation、journal 和 Reconciler
   共同表达可恢复状态。
6. 诊断是证据模型：无受控 payload 响应时用 UNVERIFIED，不把 ICMP/TCP connect 冒充
   UDP PASS。
7. IPv4-only 是新增默认，不破坏历史双栈记录。
8. 中文是显示名称，稳定 ID/Service name 不随改名变化。
9. Tunnel 负载均衡 UI 关闭，底层字段和旧配置兼容保留。
10. 原生发布使用 systemd/Caddy/SQLite 和 versioned symlink，不使用容器。
