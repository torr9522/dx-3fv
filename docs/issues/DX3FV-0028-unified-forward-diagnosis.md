# DX3FV-0028: 统一转发诊断界面

## 目标

单端口转发、隧道转发和多端口转发使用同一套诊断展示模型，同时保留各业务真实的
拓扑、探测 API、Agent 来源和证据语义。

## 设计

新增共享 `DiagnosisResultView`，统一显示：

- 业务类型与规则名称。
- 总测试数、成功数和失败数。
- 未验证、未执行和进行中数量。
- 分段路径、协议、状态、目标、测量方式和失败原因。
- 延迟、丢包率、质量等级和真实测量时间。

各页面只负责把自己的后端结果映射为公共显示项：

- 单端口：前置入口到 Landing。
- 隧道转发：入口、真实中间跳与出口。
- 多端口：前置入口到 IX、IX 到 Landing、端到端业务证据。

没有改变 Tunnel/Forward 或 IX Route 的探测命令构造，也没有合并持久化表。

## 单端口诊断补全

原 `prepareForwardDiagnosis` 假定 `tunnel_id > 0`，因此 direct Forward 无法诊断。
现在 direct Forward 使用自己的 `forward_port.node_id` 作为探测来源，并只使用当前
`remote_addr` 作为目标：

- TCP：入口 Agent 执行 `TcpPing`，语义为 TCP Connect。
- UDP：没有受控 payload 回显时不执行 `UdpPing`，明确返回
  `UNVERIFIED/not_probed`，不把 UDP dial 当作业务成功。

## 安全边界

- 所有目标来自当前业务持久化配置。
- 不扫描额外端口。
- 不从开发机探测 IX 出口公网 IP。
- 合法 Agent 到当前 Landing 的业务诊断保持允许。

## 测试

- `TestDirectForwardDiagnosisUsesTCPAndLeavesUDPUnverified` 验证 direct
  `tcp+udp` 仅派发一个 `TcpPing`，不派发 `UdpPing`。
- 前端生产构建验证三个页面共享组件的类型契约。
- 真人浏览器需分别点击单端口、隧道转发和多端口诊断，并在 1440px/390px 检查布局。

## 修改文件

- `go-backend/internal/http/handler/control_plane.go`
- `go-backend/internal/http/handler/direct_forward_diagnosis_test.go`
- `vite-frontend/src/components/diagnosis-result-view.tsx`
- `vite-frontend/src/pages/forward.tsx`
- `vite-frontend/src/pages/forward/diagnosis.ts`
- `vite-frontend/src/pages/multiport.tsx`
- `vite-frontend/src/pages/single-forward.tsx`

## 回滚

切回上一 Release。无数据库 Migration；历史 IX 诊断记录不会删除，旧页面仍可读取。
