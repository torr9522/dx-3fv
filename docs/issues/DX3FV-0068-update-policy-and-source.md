# DX3FV-0068：更新机制审计、唯一更新源与默认关闭自动检查

- GitHub Issue：`torr9522/dx-3fv#12`
- 优先级：P1
- 基线：`v0.1.68-fix-entry-address-copy`
- 数据库 Schema：不变（Schema 12）
- 默认策略：不自动检查、不自动下载、不自动安装

## 根因与调用链

右下角“发现新版本”来自 `main.tsx` 的 Vite PWA Service Worker `onNeedRefresh`，不是
Panel、Agent 或 FLVX Release。`registerSW({immediate:true})` 在登录前立即注册并检查；旧
页面由旧 Worker 控制而新 Worker 等待激活时出现提示，强制刷新后页面与已激活缓存一致，
提示消失。登录只是时间上的巧合。

另外存在两条自动 Release 查询：Footer 直接请求 GitHub，设置页加载 `/system/version`
时后端请求 GitHub。后端仓库常量仍是 `Sagit-chu/flvx`，前端 fallback 是
`Sagit-chu/flux-panel`；稳定通道还会把带 `v` 前缀的正式 Tag 错分为开发版本。

## 修复

- PWA 注册和静态资源提示默认关闭；关闭时注销当前 origin 的遗留 Service Worker。
- Footer 自动 Release 检查默认关闭。设置页提供本浏览器显式开关，开启只检查和提示，
  不下载、不安装。
- `/system/version` 只返回本地版本、能力和固定仓库，不访问外网；只有管理员点击“检查
  更新”才调用 GitHub API。
- 前后端唯一产品仓库固定为 `torr9522/dx-3fv`，默认关闭第三方 GitHub 下载代理。
- 稳定通道正确识别 `v1.0.0`；GitHub 不可达、限流、无 Release 均只返回检查失败或无
  可用版本，不改变运行版本。
- 升级 POST 必须携带 UI 确认字段。native 部署不启用继承的 Docker GUI 升级 helper，
  避免使用任何上游 GHCR；正式升级继续使用 Release 内 `upgrade-native.sh`。
- 源码包中的根目录 FLVX/Docker 遗留安装入口和 Compose 模板改为明确拒绝的占位文件，
  不再查询上游 Release 或拉取上游镜像；上游出处、许可证和历史设计文档继续保留。

## 安全边界

正式原生包的升级脚本先验证包内 `SHA256SUMS` 和 `MANIFEST`，备份 SQLite/env，切换
版本后运行健康检查，失败恢复旧 symlink、数据库和环境。v1.0.0 Release 未发布旧 GUI
所需的独立 `gost-{ARCH}` Asset，因此不能宣称 Agent GUI 升级可用；Agent 维护应使用
完整原生包和 `install-agent-native.sh`，逐节点备份和验证。

## 测试与回滚

- 后端覆盖：不自动访问 GitHub、唯一 API/下载 URL、`v` 前缀稳定 Tag、403 限流、显式
  确认、native GUI 升级拒绝。
- 既有 Agent 校验测试覆盖缺少 checksum、错误 SHA256、下载失败和回滚 binary。
- 前端生产构建与真实浏览器覆盖未登录、登录、刷新、退出、清缓存、手动检查和网络失败。
- 回滚本提交与对应发布包即可；没有 Migration，不改变数据库、Agent 或 Runtime。
