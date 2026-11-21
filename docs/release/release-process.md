# 发布流程

冻结后仅允许发布阻断、安全、数据完整性或真实业务回归修复。先完成 Issue、测试、
Commit 和 Annotated Tag，再由 `scripts/build-native-release.sh` 从 clean tagged HEAD 构建。

校验二进制/源码 SHA256、Manifest、Schema、SBOM、许可证、Secret scan；在隔离目录进行
全新安装、旧版升级、真实回滚、再次升级；最后在测试站执行桌面/手机和真实业务回归。
GitHub Release 与正式 `v1.0.0` 只能在用户明确确认后创建。正式源码 Tag 构建出的
Manifest 必须记录该 Tag 的 Commit/Tree；制品精确 SHA256 在后续 publication 提交中
加入 `main` 安装器，避免制品 Manifest 与同一提交内硬编码制品哈希形成自引用。
