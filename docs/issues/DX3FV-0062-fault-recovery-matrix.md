# DX3FV-0062：故障注入与恢复矩阵

- GitHub Issue：`torr9522/dx-3fv#6`
- 优先级：P1
- 基线：`v0.1.64-fix-full-pool-allocation-scan`
- 验证 Tag：`v0.1.65-verify-fault-recovery-matrix`
- Schema、协议和产品代码：不变
- 结论：已授权测试环境内可安全执行的故障矩阵 PASS；边界和未使用的旁路接口如实披露

## 基线与保护

注入前使用 SQLite `.backup` 保存数据库，SHA256 为
`488cc4f678fe27db8a9e63fd1d927a6a761285e0c4b8ae250a19459de48f6ee5`。基线为
Schema 12、SQLite `ok`、四节点 ONLINE、7 条 active Route、20 条 applied Runtime、
20 个 active Lease、5432 个 flow receipt。全部 IX SSH 都使用广州跳板、独立 Host Key
和严格校验；没有修改 Provider NAT、防火墙、原版 FLVX 或生产系统。

## Panel 与 SQLite

| 故障 | 注入结果 | 恢复结果 |
|---|---|---|
| Panel 正常重启 | PID 变化，启动窗口一次本机 health connect failure | health、Schema 12、SQLite 和四节点恢复 |
| Panel `SIGKILL` | systemd 记录 signal failure，restart counter 变为 1 | 3 秒后新 PID；7 Route/20 Runtime 收敛 |
| SQLite 短暂不可用 | 停 Panel 后把精确数据库文件临时设为 mode 000；新进程确定性启动失败 | 恢复 owner 和 0640 后启动；在线库和备份均 `integrity_check=ok` |

Panel 重启日志保留了 Agent 尚未全部重连时的真实 `节点不在线`，以及后续自动恢复；没有
把启动瞬间的失败隐藏为始终可用。

## Agent、GOST 与 Listener

- 广州、前海、QIX、盲盒逐个正常重启，PID 均变化，Listener 数分别恢复到
  17、8、1、2。重启后七场景真实 payload 7/7 通过。
- 广州 Agent 主进程执行 `SIGKILL` 后 systemd 自动拉起，新 PID、restart counter 1、
  17 个 Listener 恢复；前海 TCP/UDP、QIX TCP/UDP、盲盒 TCP 随后全部通过。
- 前海 Agent 停止时 Panel 状态为 OFFLINE，Listener 为 0，真实 TCP 和 UDP 均失败；
  启动后 ONLINE、8 个 Listener、8 条 applied Runtime 和 TCP/UDP 全部恢复。
- 本项目 GOST Runtime 内嵌在 Agent 进程，没有独立外部 GOST PID，因此 Agent SIGKILL
  同时覆盖 Runtime 主进程故障，QIX Agent 重启同时覆盖共享 Relay Runtime 重启。
- Agent lifecycle 集成测试证明失败配置 reload 不会丢失已确认 Listener。没有绕过 Panel
  所有权调用未公开控制端点来强删单个 Service；单 Service 真实失败由下面的 bind
  conflict 覆盖。

## 多端口部分失败

Reconciler 的真实安全顺序是先 IX、后公网 Entry：

- IX 离线创建：IX Runtime 为 `error/applied=0`，Entry 为 `pending/applied=0`，没有先
  暴露公网入口；IX 恢复后两段自动 active。
- GZ 离线创建：IX 内部 Runtime active，Entry 为 `error/applied=0`；GZ 恢复后收敛。
- IX 离线删除：Route 为 `cleanup_pending`，两 Lease 为 `pending_release`；恢复后两段
  `absent`、Lease `released`。
- GZ 离线删除：IX 已 absent，Entry error，Route 为 `cleanup_pending`；恢复后完整删除。
- 手动指定已占用的 `41004` 被明确拒绝，没有 Runtime 或 Lease 残留。
- 自动候选 `58156` 被真实的外部 Listener 占用时，系统记录 bind conflict、释放旧 Lease，
  自动改用 `58157` 并 active；临时 blocker 和测试 Route 随后清理。
- 同时竞争创建由 DX3FV-0061 的五 Route 真实 CRUD 和 12 worker 仓储测试覆盖。
- Landing 编辑期间停止 GZ，API 返回失败、journal 为 `rollback_pending`；节点恢复后
  journal 为 `rolled_back`、旧 Route active、真实 TCP payload PASS，active migration 为 0。

## QIX 共享 Relay 隔离

在现有两个 Tunnel Forward 之外创建第三个隔离测试 Forward：创建 payload PASS；暂停
后自身 payload 失败而兄弟 QIX TCP/UDP PASS；恢复、编辑后自身 PASS；删除后两个兄弟
仍 PASS。QIX Relay Listener 在测试前后均为 1，没有因删除一个消费者而释放共享 Relay。

## Agent 升级失败与回滚

盲盒使用隔离的失败候选目录和 `/bin/false` 模拟升级二进制。服务按预期启动失败；恢复
旧 `current` symlink 后，原二进制 SHA256 不变、2 个 Listener 恢复、真实 TCP payload
PASS。失败候选目录已删除，旧 Release 和证据保留。

## 最终状态

最终七场景真实回归 7/7 通过。四节点 ONLINE；active Route/Runtime/Lease 仍为
7/20/20；所有 DX0062 临时 Route 和 Forward 为 0；incomplete migration 为 0；flow
receipt 从 5432 增至 5514；SQLite integrity 为 `ok`。历史 Route、Runtime、Lease 和
migration 记录按审计模型保留，不进行破坏性清理。

## 自动化与证据

后端定向测试覆盖 Runtime 生命周期、部分删除、自动/手动冲突、编辑回滚、共享资源离线
保留与 reconnect repair；全部通过。`go-gost/tests/lifecycle` 通过。

本地受控证据目录：`DX3FV-0062/`。

- `matrix-summary.json`：`551eed8580cab548c65ee3f82bfee83b23f547b62626ceabc7febc788d13da3f`
- `panel-baseline.txt`：`a5120fc7e5a2d8199c724df81429a38fc1cc4b7abd6dacf4843bf787279a4b7d`
- `agent-restarts-payload/result.json`：`bf92bf0d507ce79a8707f59461ca941185ba708733f81df3ff7ff006cc3de722`
- `final-recovery-payload/result.json`：`b2fe93d9682b02ce2cf03ab99c767aff02170e8a10b15d48ac8c3721a6a4a00c`
- `final-recovery-payload/events.ndjson`：`fb3216129d77048b355f40f151e9b6d780c1842ab6be0a03e377b272706518f8`
- `final-panel-state.txt`：`73a59d049e55fd4e7b7c55abd4961f930ba8d3c8907109ef28e9b9e31eff069c`

证据不包含 SSH 密码、GitHub Token、Agent Secret、管理员密码或 Shadowsocks URI。

## 回滚

本 Issue 只提交验证文档，不修改产品二进制或 Schema，无产品代码可回滚。所有注入均已
在现场恢复；若需要恢复测试 Panel，可使用注入前 SQLite 备份和已保留的
`v0.1.62-migrate-react-router-7` Release。不要删除故障时间线和历史状态记录。
