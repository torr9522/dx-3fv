# Release Candidate 已知问题

## 前海 UDP 会话陈旧风险

持续回归曾两次出现前海 UDP 业务 timeout；同一 Landing 的 QIX UDP 正常，只重启
广州入口 Agent 后前海恢复。现有证据指向入口侧 UDP Association/会话陈旧方向，
但没有定位到精确代码根因。最终 v0.1.48 单轮回归通过，不能据此关闭历史风险。

本 RC 不对 GOST 核心 Runtime 做推测性修改。出现时应保存时间线、Runtime/Agent
状态和端口证据，再按运维指南恢复。

## 稳定性覆盖有限

已完成真实 TCP/UDP/Relay 业务与重启恢复回归，但未形成小时级持续流量、极限高并发
或所有部分失败组合的完整证明。Release Candidate 不应描述为无限负载或长期无故障。

## 前端依赖延期

当前 audit 的 13 条公告位于 Tailwind/PostCSS 构建链，不进入浏览器 bundle；另外
两条 React Router moderate 公告的修复要求迁移到 Router 7.18，涉及本项目未使用的
SSR error hydration 和用户可控导航目的地路径。盲目主版本升级的回归风险高于已
确认可达风险。详情见 RC 安全审计，正式发布前仍应复核新的安全公告。

## 历史 Secret 候选项

完整 Git 历史包含三个已删除上游 GOST 配置中的 32 位十六进制 secret/password。
当前 Tree 和发布归档不包含这些值，但仓库无法证明其从未使用或已经失效。因禁止
重写上游历史，公开推送前必须由所有者确认失效或轮换并接受历史保留风险。

## 平台范围

原生发布主要在 Debian 12 amd64、systemd、Caddy、SQLite 上验证。其他发行版、
架构、PostgreSQL 大规模数据和多 Panel 副本需要单独验收。WebAuthn challenge 保存在
进程内存，多 Panel 副本需要共享会话设计。
