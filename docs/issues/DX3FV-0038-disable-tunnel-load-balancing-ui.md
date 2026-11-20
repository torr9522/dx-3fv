# DX3FV-0038：隧道负载均衡审计及关闭操作入口

- 基线：`v0.1.41-fix-forward-page-scope`
- 范围：FLVX 3.0.1-rc16、DX-3FV Panel/Agent、内嵌 go-gost/x
- 数据库变更：无
- Runtime 能力删除：无

## 1. 审计结论

编辑 Tunnel 页面中的四个中文选项不是展示别名，而是写入
`chain_tunnel.strategy` 并参与 GOST hop selector。它们只在同一跳或出口存在多个候选
节点时改变选路；单候选时结果相同。

| UI   | 存储值  | Runtime 行为                                | 自动切换                                       |
| ---- | ------- | ------------------------------------------- | ---------------------------------------------- |
| 主备 | `fifo`  | 选择排序后的第一个可用节点并持续使用        | 实际拨号失败后由 FailFilter 排除，尝试后续节点 |
| 轮询 | `round` | 每次选择依次轮换候选节点                    | 失败候选同样被 FailFilter 临时排除             |
| 随机 | `rand`  | 按权重随机选择，缺省权重为 1                | 失败候选同样被 FailFilter 临时排除             |
| 最优 | `best`  | Panel 对候选出口排序，再向 GOST 下发 `fifo` | 周期质量探测连续确认后重排 chain               |

GOST 还支持 `hash`，但原 Tunnel UI 没有暴露。未知 strategy 在 Fork parser 中会回退
到 round；本 Issue 不改变这一底层兼容行为。

## 2. 数据、API 与调用链

1. `go-backend/internal/store/model/model.go` 的 `ChainTunnel.Strategy` 映射
   `chain_tunnel.strategy`，每个 entry/middle/exit 候选行各自保存。
2. `go-backend/internal/http/handler/mutations.go` 的
   `prepareTunnelCreateState` 读取 `inNodeId/chainNodes/outNodeId[].strategy`；空值历史缺省为
   `round`。
3. `replaceTunnelChainsTx` 调用
   `go-backend/internal/store/repo/repository_mutations.go` 的
   `CreateChainTunnelTx` 保存；`ListTunnels` 在
   `repository.go` 中把原值返回相同数组，编辑表单此前可以覆盖它。
4. `buildTunnelChainConfig` 将候选节点生成 GOST hop，并把策略放进
   `selector.strategy`；`maxFails=1`、`failTimeout=10m`。
5. Agent 通过现有 WebSocket `create/update chain` 命令接收整份 GOST 配置。Agent 不重新
   解释中文策略，也不另存一个策略字段。
6. `go-gost/x/config/parsing/selector/parse.go` 把 `fifo/ha`、`round/rr`、
   `random/rand` 分别映射为 FIFO、RoundRobin、Random selector。
7. `go-gost/x/selector/filter.go` 的 FailFilter 根据 Node Marker 过滤失败候选；
   `go-gost/x/chain/route.go` 在拨号失败/成功时 Mark/Reset。因此这属于业务拨号触发的
   被动健康状态，不是独立端口扫描。

## 3. “最优”策略和自动探测

`best` 是 DX-3FV/冻结参考实现中的 Panel 扩展，不是 GOST parser 原生策略。
`tunnel_quality_prober.go` 按配置周期检查启用的 Tunnel；仅当出口多于一个且策略为
`best` 时，`probeBestExitOwners` 对候选评分。连续三轮满足优势门槛并通过冷却期后，
`applyBestExitChainOrder` 重排候选并重新下发。`runtimeTunnelStrategy` 将 `best` 转为
GOST 可执行的 `fifo`。

普通 Tunnel 质量监控和手动诊断是独立功能，关闭负载均衡操作入口后仍保留。旧
`best` Tunnel 继续按原配置运行；本 Issue 不静默关闭或迁移它。

## 4. 多出口、路径和单出口语义

- 多节点 middle group 或多个 exit 才形成 selector 的多个候选，策略会改变实际路径。
- `fifo` 在首选失败时可切到后续候选；`round` 和 `rand` 在正常状态也会分散连接。
- 单出口或单节点 middle group 只有一个候选，四种策略在路径选择上没有差异；
  FailFilter 对单候选明确 fail-open，避免永久过滤唯一节点。
- entry 列表表示多个独立入口所有者，不是同一个 GOST hop 的出口负载均衡。
- Forward 表还有独立的 `forward.strategy`，用于多个 Landing 目标；不属于本 Issue，
  不作修改。

测试环境基线只有一个 Type-2 QIX 共享 Relay。entry/exit 都各一个候选，存储策略均为
`round`、传输均为 `tls`；关闭入口前已记录，后续编辑必须保持这些值。

## 5. TCP、UDP 与传输边界

负载策略只选择下一跳候选，不决定业务协议。Type 1 的 TCP/UDP 由普通 Forward
Service 决定；Type 2 的 Relay connector/handler 继续在一个 Relay Service 中承载逻辑
TCP 和 UDP。`tcp/tls/wss/mtls/mwss/mtcp/kcp` 仍是节点间外层传输，TLS 不是业务
协议。本 Issue 不修改 `DX3FV-0026` 的能力矩阵，也不改 QIX 共享 Relay。

## 6. 操作层关闭方案

- 新建和编辑表单移除主备、轮询、随机、最优四个下拉选项。
- 新 Tunnel 的每个 middle hop 和 exit 只允许一个节点，内部保存 `fifo`，形成固定路径。
- 编辑旧单节点 Tunnel 时保留原 strategy，不批量改写。
- 编辑旧多节点 Tunnel 时保留候选顺序、strategy 和传输；多节点组显示为只读兼容配置，
  其他名称、流量、状态、诊断参数仍可编辑。
- API、数据库列、Panel best manager、Agent 和 GOST selector 保留，保证旧配置继续运行和
  版本回滚。

## 7. 测试与回滚

自动化覆盖：固定策略常量、新建多出口/多节点跳拒绝、旧多节点编辑兼容、源码中不再
挂载四个操作选项。Go 全量测试和前端生产构建必须通过。

2026-11-20 自动化结果：上述前端契约测试及既有 Landing/页面边界/折叠测试 9/9
通过；TypeScript/Vite production build 通过；Agent、内嵌 go-gost/x 全量测试通过；
Panel Handler 与 Repository 套件通过。首次 Panel 全量运行在既有 suite-scope 风险后
返回失败，`tests/contract` 隔离重跑通过，按 `DX3FV-0001` 继续跟踪，不归入本功能提交。

真人浏览器在桌面 1440px 和手机 390px 验证：新建默认 TCP、无负载策略控件、旧 QIX
Tunnel 编辑后数据库策略/传输不变、共享 Forward TCP/UDP/TCP+UDP 与删除隔离不退化。

回滚：切换 Panel `current` symlink 到
`v0.1.41-fix-forward-page-scope` 对应 release 并重启。无 Migration、无数据重写，旧版可
直接读取所有 strategy。
