# DX3FV-0061：高并发与资源极限验收

- GitHub Issue：`torr9522/dx-3fv#5`
- 优先级：P1
- 基线：`v0.1.63-verify-hour-scale-traffic`
- 修复 Tag：`v0.1.64-fix-full-pool-allocation-scan`
- Schema、Agent 协议和 GOST Runtime：不变
- 结论：100 并发目标已验证；并发 CRUD 暴露并修复全端口池穷举阻塞

## 分级数据面测试

每个并发级别运行两轮，不重试失败请求。覆盖单端口 Forward、前海多端口 TCP、
前海 UDP、前海 TCP+UDP、QIX 两个 Forward 共享 Relay、原版 Type-2 Tunnel 的单个
Forward，以及多个 PathProfile 聚合。测试从 1、10、25、50 递增到 100；达到发布
目标后停止，没有为了追求更大数字继续压占共享测试节点。

| 并发级别 | 通过/总数 | 说明 |
|---:|---:|---|
| 1 | 14/14 | 七场景两轮全部通过 |
| 10 | 140/140 | 全部通过 |
| 25 | 348/350 | 前海 UDP 第二轮两个单包超时 |
| 50 | 698/700 | 前海 UDP 98/100；单端口 TCP P95 5.420 秒，首次明确延迟退化 |
| 100 | 1396/1400 | 前海 TCP 199/200、QIX 共享 Relay 197/200；其余场景全部通过 |

100 并发的四次失败均为访问外部 Cloudflare 目标时的 TLS handshake error，前后请求
成功，UDP 为 200/200，节点、Listener 和进程未中断。最大已验证并发是 100；这不是
硬件容量上限，也不能描述为无限并发或零错误 SLA。

143 份资源快照中，Panel 最大 RSS 66.5 MiB、FD 16；广州 Agent 最大 RSS 90.5 MiB、
FD 90、TCP socket 76；前海 Agent 最大 RSS 75.9 MiB、FD 20；盲盒最大 RSS 53.8 MiB、
FD 9；QIX 最大 RSS 64.3 MiB、FD 18。四节点离线样本、进程重启、SQLite 错误和采集
错误均为 0，压力结束后 FD 恢复到基线附近。部署未公开 Go profiling，因此 goroutine
精确值仍为 `NOT_EXPOSED`。

## 并发 CRUD 根因

首次同时创建五条隔离的前海 TCP Route 时，五个 API 请求均在约 124.74 秒由 Caddy
返回 502，但后端随后完成了 Route 18--22。同期无关 `/flow/upload`、`node/list` 和
`user/package` 请求被阻塞 8--120 秒。Go HTTP `WriteTimeout` 为 2 分钟，因此 Handler
完成前连接已被关闭。

源码与现场 SQL 时间线确认：SQLite 使用 `SetMaxOpenConns(1)`；
`AllocateIXPorts` 在事务中锁定 Pool；`full` Pool 的候选生成会从 `1..65535` 查询并
物化全部可用 Resource，而调用者实际上只消费第一个候选。五个创建事务串行执行这段
穷举，使单连接 SQLite 上的全部 Panel 请求排队。根因不是 Caddy、HTTP timeout、Agent
或 GOST；扩大 timeout 只会隐藏阻塞。

## 修复

`ixPortCandidates` 在持有 Pool 行锁的事务内找到首个未占用的 full-pool Resource 后
立即返回。NAT Pool 逻辑不变；手动端口、TCP/UDP 独立所有权、SSH 22 排除、Lease
唯一约束和 Pool `next_port` 轮转不变。SQLite 由单连接事务串行保护，PostgreSQL 继续
使用 `FOR UPDATE` 的 Pool 行锁保护并发分配。

自动化回归新增两次自动分配检查，要求端口不同且只物化两个 Resource，而不是约
65K 个。现有 12 worker TCP+UDP 并发分配测试连续运行 20 次通过；后端全套测试通过。

## 真实修复验证

修复 canary 上五条并发创建在 0.697 秒完成，全部 HTTP 200；同一创建窗口三个
`node/list` 请求为 47.7--59.4 ms，无 5xx。五条 Route 的 Entry/IX 端口均唯一。
创建后及并发编辑后各执行五次真实 Shadowsocks TCP 16 KiB payload，共 10/10 通过。
五条并发删除全部成功；Route 为 `deleted/deleted`，十个 Runtime lane 为
`deleted/absent`，相关 Lease 全部 `released`，五个 migration journal 均为
`complete`，SQLite integrity 为 `ok`，Schema 12，Panel 未重启。

第一次修复后 payload 复测保留两类测试端问题：外部 HTTPS 目标同时握手失败，以及
同时启动五个本地 shadowsocks-libev 客户端受本机测试资源限制。最终验证改用同一外部
目标的明文 HTTP 16 KiB payload，并逐 Route 启动隔离客户端；没有改变服务端、Route
或业务验收语义，也没有重试失败 payload。

## 证据

本地受控证据目录：`DX3FV-0061/`。

- `formal-1-25/result.json`：`012a70c6ddb21ad09ef04972cedd15e215888d00bf1c1362f74e2abe9b9ba42c`
- `formal-50/result.json`：`0a5f2840dc22c7fe34475747a458429aa0c59c8469f4ba2147b9fc73b1e5c622`
- `formal-100/result.json`：`e63b289f7578b6f9b91f67e339b76db935eaf1b86e0695e157e30ee34917b582`
- `formal-100/events.ndjson`：`e84e1158a142880f1d0a9746a5e0d393c1301bd6881251327e3b094bd14b3e87`
- `formal-100/resources.ndjson`：`700df3d8d9b7198362e4b50bb3292bd7c78b1a0e04c3c78e8da0b54fba080931`
- 修复前 CRUD：`2040defc44f16d237adf26c25a3ecc06ee5ef9715420d5bfa486b979b2d07b7a`
- 修复后完整 CRUD：`42ed6f9df6b71a860628a2c986cacbce7f4a6b47295e5ae89815443f9294582d`

证据不包含 Shadowsocks URI、密码、Agent URL、SSH 凭据或 GitHub Token。

## 回滚

回滚到 `v0.1.63-verify-hour-scale-traffic` 的 Panel 二进制即可；Schema 和数据无需回滚。
回滚会恢复已确认的 full-pool 穷举阻塞，因此只能作为短期应急，并应避免并发自动创建。
不要删除测试证据或已完成的 migration/Lease 历史记录。
