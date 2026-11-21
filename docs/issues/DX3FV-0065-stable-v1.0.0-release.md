# DX3FV-0065：准备 v1.0.0 稳定版

- GitHub Issue：`torr9522/dx-3fv#9`
- 优先级：P0 发布阻断
- 基线：`v1.0.0-rc.4`
- Schema：12

## 范围

从已完成安装、升级、回滚、真实业务、持续流量、并发和故障恢复验收的 RC4 生成首个
稳定版。只补充正式 Release Notes、网络依赖边界、版本元数据、制品和安装器固定映射；
不增加功能，不改变数据库 Schema、Agent 协议或 GOST Runtime。

正式源码 Tag 与制品 Manifest 使用同一个 Commit/Tree。由于 tar 包的 Manifest 包含
Commit/Tree，不能在同一提交内再硬编码该 tar 包自身 SHA256；精确制品哈希与安装器
映射使用独立的后续 publication 提交，与 RC4 的可审计发布模式一致。

## 阻断检查

- 公开仓库 worktree clean、唯一 remote、local/main 与 origin/main 一致。
- RC4 Annotated Tag、Commit、Tree、Manifest 和外部 SHA256 一致。
- 包内逐文件 SHA256、SBOM、LICENSE、NOTICE 完整。
- publication history、Tree、CI 日志和发布归档通过 Gitleaks。
- 唯一归档候选为已精确 allowlist 的前端 localStorage 键名，不是 Secret。
- Native CI、两站升级/健康检查、正式制品远程回下载校验全部通过后才发布。

## 回滚

正式源码无运行逻辑变化。部署回滚使用发布包自带 `rollback-native.sh` 和升级前备份；
Git 历史不重写、不移动 Tag，安装器可显式选择 `v1.0.0-rc.4`。
