# 发布检查清单

- clean worktree，HEAD 有唯一 Annotated Tag，Commit/Tree 已记录。
- 后端、Agent、go-gost/x tests 与前端 build 通过。
- Schema、Panel/Agent 版本和 GOST fork 对应同一源码。
- 当前 Tree、完整 publication 历史、源码包、二进制包 Secret 扫描通过。
- SHA256、SBOM、LICENSE、NOTICE、第三方声明完整且无绝对构建路径。
- Debian 12 clean install、登录、重启恢复通过。
- 升级/回滚恢复数据库、Agent、Listener、TCP/UDP/Relay 和流量。
- 1440px/390px 浏览器无 5xx、JS error 或水平溢出。
- Known Issues 与淘汰 RC 如实披露。
- 只向 `torr9522/dx-3fv` push；不 force push、不覆盖 Tag、不创建稳定版。
