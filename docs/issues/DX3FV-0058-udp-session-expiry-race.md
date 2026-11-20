# DX3FV-0058：UDP 会话到期与入队竞态

- GitHub Issue：`torr9522/dx-3fv#2`
- 优先级：P0
- 基线：`v0.1.60-fix-multiport-create-scope`
- 发布 Tag：`v0.1.61-fix-udp-session-expiry`
- 数据库 Schema：不变（Schema 12）
- Panel/Agent 协议：不变

## 现象与根因

前海 UDP 曾在空闲一段时间后超时，而同一 Landing 的 QIX UDP 正常；重启广州入口
Agent 后恢复。该恢复动作只能定位到入口 Runtime 生命周期方向，不能证明 NAT 是根因。

`go-gost/x/internal/net/udp/listener.go` 的旧 `WriteQueue` 同时选择向缓冲 `rc` 写入和
读取已关闭的 `closed` channel。连接关闭且队列仍有容量时两个分支均 ready，Go 会随机
选择，datagram 可能进入已经没有消费者的旧 session。idle scanner、首包并发创建和旧
session 删除 replacement 还存在相互独立的竞态窗口。

确定性回归测试 `TestClosedConnNeverAcceptsQueuedDatagram` 在修复前以 `-count=10`
全部失败，通常第 0--3 次就观察到已关闭连接接受 datagram。当前 go-gost/x master
`f810ee231df8b46fcf8e76fd571d40124f2e3a64` 仍保留相同无锁结构。FLVX
`bb48ab00bd276460d618c9be0b2b2be5639746cb` 只在入队后调用 `SetIdle(false)`，没有消除
关闭与入队竞态。

## 修复

- `Close`、idle expiry 与 `WriteQueue` 使用同一把锁；检查到 closed 后绝不再入队。
- pool 使用 `LoadOrStore`/`CompareAndSwap`，同一 peer 并发首包只能发布一个新 session。
- idle scanner 使用 `CompareAndDelete`，旧 session 到期不会删除已经替换的新 session。
- datagram 恰遇到旧 session 关闭时，只在新 session 重投一次；不增加通用重试。
- 所有未入队 buffer 都归还 bufpool。
- 不改变 UDP TTL、业务 timeout、Panel 状态语义、Agent 协议或 GOST Service 配置。

## 自动化测试

- `TestClosedConnNeverAcceptsQueuedDatagram`：修复后 100 次 PASS。
- `TestConcurrentFirstDatagramsShareOneSession`：验证同 peer 只发布一个 session。
- `TestExpiredSessionCannotDeleteReplacement`：验证旧 session 不能误删 replacement。
- UDP race 套件 25 次 PASS。
- `go-gost/x` 全套、GOST lifecycle、backend Repository 与 Contract 全套 PASS。

## 真实业务验收

广州2号与前海IX2先部署同源码 canary，均保持 `NRestarts=0`：

- 前海两条 TCP、QIX 两条 TCP、盲盒两条 TCP 均返回 HTTP 200。
- 前海 UDP、前海 TCP+UDP 的 UDP lane、QIX UDP 各 150/150，总计 450/450。
- 完全空闲 70 秒后三条 UDP 首包全部成功；前海 UDP P99 238.663 ms，最大
  242.733 ms。
- 真实浏览器两次创建、UDP DNS payload、删除、重建循环均通过，共 10/10 payload，
  无 JS error 或 HTTP 5xx。
- 删除后 Route 审计记录为 `deleted/deleted`，四条 Runtime 为 `deleted/absent`，无活动
  Lease；广州测试端口与前海内部端口均无 Listener。
- SQLite integrity `ok`；原有七条业务保持 active。

证据位于本地发布证据目录 `DX3FV-0058`；结果文件 SHA256：

- `udp-session-lifecycle.json`：
  `e6c275b8b79406515c17648b41fb92678a4092f07e272e25d567dd70b53b38e5`
- `browser-recreate/result.json`：
  `452cd0154ba61f0cd826ae6360bbfcb2605d91339bbe1c116f68fbbfdd1b12f6`

这组结果关闭已复现的 session expiry race，但不替代 DX3FV-0060 的小时级持续流量测试。

## 回滚

回滚本提交并把广州2号、前海IX2 Agent symlink 切回
`v0.1.46-feature-forward-navigation-collapse`，随后依次重启 Agent 并验证 Listener、TCP、
UDP 和兄弟业务。没有 Migration；数据库与现有 Route、Lease、Owner 无需回滚。
