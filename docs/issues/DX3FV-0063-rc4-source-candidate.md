# DX3FV-0063：v1.0.0-rc.4 源码候选冻结

- GitHub Issue：`torr9522/dx-3fv#7`
- 优先级：P1 发布准备
- 基线：`v0.1.65-verify-fault-recovery-matrix`
- 目标 Annotated Tag：`v1.0.0-rc.4`
- Schema：12

## 范围

冻结 DX3FV-0057 至 DX3FV-0062 已验证的源码和文档，新增 RC4 Release Notes。此步骤不
修改产品行为、数据库、Agent 协议或 Runtime。正式 `v1.0.0` 与 GitHub Release 均不在
授权范围。

## 构建顺序

源码 Tag 必须先保持不可变，随后从干净 Tag 使用 `scripts/build-native-release.sh` 和
固定 Go 1.26.11 构建。归档 SHA256 只有构建后才能确定，因此安装器映射和二进制归档
由后续独立 publication Issue 提交，禁止回写或覆盖本 Tag。

## 验证

Tag 前置条件：六个 Phase 5 Issue 已关闭；后端、前端、Agent lifecycle、真实持续流量、
100 并发和故障矩阵均已有独立证据；源码 Tree 和完整 sanitized history 的 Secret 扫描
为 0 findings；工作区只包含本 Issue 发布文档。

## 回滚

源码无 Migration，回滚到 `v0.1.65-verify-fault-recovery-matrix` 即可。Annotated Tag
不可覆盖或删除；若正式构建失败，应保留失败 Tag 并创建新的候选，不能改写 RC4。
