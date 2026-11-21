# DX3FV-0060：小时级持续流量验收

- GitHub Issue：`torr9522/dx-3fv#4`
- 优先级：P1
- 基线：`v0.1.62-migrate-react-router-7`
- 验证 Tag：`v0.1.63-verify-hour-scale-traffic`
- 数据库 Schema、Panel/Agent 协议和 Runtime：不变
- 结论：小时级连续运行主链 PASS；观察到的单包丢失和外部 TLS 失败如实保留

## 方法

先完成 60 秒基线，再在同一未重启进程窗口并行运行 3660 秒。TCP 每 15 秒通过真实
Shadowsocks 下载 64 KiB；UDP 每 2 秒通过真实 Shadowsocks 发送 DNS payload。覆盖：

- 前海 TCP、UDP、TCP+UDP；
- QIX 两条 Forward 复用真实 Type-2 Tunnel 共享 Relay，均运行 TCP+UDP；
- 盲盒自动和手动两条 TCP-only 业务；
- 单端口手动 TCP+UDP 和自动 TCP 业务。

每分钟经广州跳板采集 Panel 与四节点的 CPU、RSS、FD、线程、socket、重启计数、
SQLite integrity、flow receipt、节点状态和逐业务累计流量。部署没有公开 Go profiling，
因此 goroutine 精确值为 `NOT_EXPOSED`；未使用 SIGQUIT 或公开 pprof 制造该指标。

实际运行 `3666.248` 秒，记录 10697 个 payload 事件和 62 份资源快照。测试不重试失败
payload，也未重启 Panel、Agent 或 GOST 来恢复业务。

## 数据面结果

| Lane | 通过/总数 | 成功率 | P50 | P95 | P99 | 最大 |
|---|---:|---:|---:|---:|---:|---:|
| 前海 TCP | 244/244 | 100% | 1858 ms | 1985 ms | 2449 ms | 3521 ms |
| 前海 UDP | 1748/1749 | 99.943% | 233 ms | 238 ms | 270 ms | 292 ms |
| 前海混合 TCP | 244/244 | 100% | 1861 ms | 1931 ms | 2242 ms | 2934 ms |
| 前海混合 UDP | 1749/1749 | 100% | 379 ms | 385 ms | 417 ms | 438 ms |
| QIX tri TCP | 244/244 | 100% | 1878 ms | 2044 ms | 2381 ms | 2642 ms |
| QIX tri UDP | 1743/1749 | 99.657% | 400 ms | 406 ms | 439 ms | 459 ms |
| QIX dmit TCP | 237/244 | 97.131% | 1449 ms | 1744 ms | 2029 ms | 2203 ms |
| QIX dmit UDP | 1749/1749 | 100% | 229 ms | 231 ms | 267 ms | 289 ms |
| 盲盒自动 TCP | 244/244 | 100% | 1879 ms | 1977 ms | 2569 ms | 4936 ms |
| 盲盒手动 TCP | 239/244 | 97.951% | 1480 ms | 1806 ms | 2147 ms | 2935 ms |
| 单端口手动 TCP | 244/244 | 100% | 1879 ms | 1988 ms | 2408 ms | 3329 ms |
| 单端口手动 UDP | 1749/1749 | 100% | 383 ms | 392 ms | 423 ms | 444 ms |
| 单端口自动 TCP | 244/244 | 100% | 1876 ms | 1983 ms | 2203 ms | 4838 ms |

协议汇总：UDP 8738/8745（99.920%），TCP 1940/1952（99.385%）。UDP 和 TCP 的最大
连续失败均为 1。

前海 UDP 唯一超时发生前 1.9 秒请求成功，后 5.3 秒同一 lane 自行恢复；QIX tri 的
6 次 UDP 超时也都被同一 lane 的成功请求夹住。没有出现历史上持续到重启 Agent 才
恢复的陈旧会话模式。12 次 TCP 失败均为 curl 访问外部
`speed.cloudflare.com:443` 时在 TLS 握手阶段报告 `SSL_ERROR_SYSCALL`；相同 lane 的
前后请求成功，Agent、Listener 和共享 Relay 没有中断。错误仍计入成功率，不能把本
结果描述为零丢包或外部 Landing SLA。

## 资源与累计流量

| 进程 | RSS 首/末/最大 | FD 首/末/最大 | 线程范围 | 重启 |
|---|---:|---:|---:|---:|
| Panel | 66.1/66.5/66.5 MiB | 15/14/16 | 9 | 0 |
| 广州 Agent | 55.8/58.2/58.2 MiB | 35/37/37 | 8 | 0 |
| 前海 Agent | 53.3/53.6/53.6 MiB | 23/25/25 | 7-8 | 0 |
| 盲盒 Agent | 51.8/53.8/53.8 MiB | 8/9/10 | 7 | 0 |
| QIX Agent | 46.5/49.2/49.2 MiB | 12/12/12 | 7-8 | 0 |

- 62/62 快照采集成功，分钟间隔 56.940--63.254 秒；节点离线样本为 0。
- SQLite 每次 `integrity_check` 均为 `ok`；数据库增长 712704 B。
- flow receipt 计数从 2468 单调增至 5146，增加 2678。
- 9 条业务在全部 61 个相邻快照区间都至少一个方向增长；负增量 0，双向停滞 0。
- QIX 两条共享 Relay Forward 分别累计，删除/隔离语义未在本 Issue 注入故障，由
  DX3FV-0062 单独验证。
- 客户端值是应用有效载荷；Panel 值包含 Shadowsocks、TCP/UDP、Relay 和上报周期
  开销，因此不要求逐字节相等。各 route 的 Panel 增量均大于对应成功 payload，且数量级
  与事件总量一致，没有跨 route 合并或双倍累计迹象。
- RSS 在负载开始后出现小幅 Go heap 台阶，FD、socket 与线程保持窄区间，没有小时窗口
  内的单调泄漏证据；这不等于无限期内存稳定性证明。

## 证据

本地受控证据目录：`DX3FV-0060/hour-61m/`。

- `analysis.json`：`c901e968c861663d8b68d6c9a55a03534914d1a3d25b2f5f72052b806593c9cb`
- `result.json`：`7878bfce91b5d6122cf8f2db9c161911971d92626e21b9a22d7def008af56c40`
- `payload-events.ndjson`：`ae9e89079a3457ab771e711aa6ebd21b31a1519f675f2a28a483db0a7209cd50`
- `resource-snapshots.ndjson`：`6f2d96cf57b5438fad0cd67e465d22c272198590e072f92359504fd39a38ae4d`

证据不包含 Shadowsocks URI、密码、Agent URL、SSH 凭据或 GitHub Token。

## 回滚

本 Issue 只提交测试结论和文档，不修改二进制、数据库或部署。回滚文档提交即可；不要
删除原始证据。测试结束后所有临时 Shadowsocks 客户端和临时凭据目录均已清理。
