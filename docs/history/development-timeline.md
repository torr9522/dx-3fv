# 开发时间线

- 基线：冻结 FLVX rc16，保留原 Forward、Tunnel、Node、用户与 Agent/GOST 合同。
- Phase 1/2：建立 IX PortPool/Resource/Lease/Owner、PathProfile、IXRoute、两段 Runtime、
  Reconciler、迁移 journal、诊断和累计流量；坚持不创建 synthetic Tunnel。
- Phase 2.5：形成转发一级导航及单端口/隧道/多端口三个页面，审计 Tunnel 协议，统一
  诊断、Landing 输入、前置文案和分组折叠。
- Phase 2.6：修正 Echo 误报，默认 IPv4-only，验证中文名称与导航折叠，完成 Agent
  兼容和前端依赖收尾。
- Phase 3：功能冻结，整理原生安装/升级/回滚、许可证、Secret 和 SBOM。RC1 因 Go
  SBOM 绝对路径淘汰，RC2 因构建模板错误淘汰，RC3 通过构建与测试站演练。
- Phase 4：只建立 sanitized 公开历史、独立知识库和 Debian 12 一键安装验收。
- Phase 5：修正多端口新增业务 PathProfile 作用域，关闭 UDP session expiry 入队竞态，
  迁移 React Router 7，并以 10697 个真实 payload 完成 61 分钟持续流量验收；真实单包
  丢失与外部 TLS 失败保留为发布证据，不宣称零丢包 SLA。

逐项根因和 Tag 见 `docs/issues/` 与 `docs/history/release-history.md`。
