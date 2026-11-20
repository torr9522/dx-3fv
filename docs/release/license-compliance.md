# 许可证合规审计

本文件记录工程事实，不构成法律意见。

## 项目来源

- flux-panel 上游声明 Apache-2.0，原始声明及许可证保存在 `NOTICE` 与
  `LICENSE-APACHE`。
- FLVX 对上游进行了后端、前端和 Agent/GOST 集成改造，并声明修改部分为 GPLv3。
- DX-3FV 从冻结的 FLVX `3.0.1-rc16` 源码继续修改，根目录 `LICENSE` 为 GPLv3。
- 内嵌 `go-gost/` 与 `go-gost/x/` 源自 MIT 项目，许可证保存在
  `LICENSES/GOST-MIT.txt`。
- 前端 bridge 含 Next UI MIT 代码，许可证保存在 `LICENSES/NEXTUI-MIT.txt`。

## ForwardX

开发期间只研究了 ForwardX 的外部行为和调用机制，没有复制其 AGPL-3.0 源码进入
DX-3FV。基于当前源码审计，DX-3FV 不因该研究变为 AGPL 项目。若未来引入 ForwardX
代码，必须重新进行来源与兼容性审查。

## 分发义务

分发 DX-3FV 二进制时应同时提供同一 Commit 的完整、可构建对应源码，保留 GPLv3、
Apache-2.0、MIT 许可证和 NOTICE，并提供构建安装信息。发布脚本为每个版本生成匹配
的 `-source.tar.gz`；二进制包包含 Manifest、SBOM 和第三方清单。

仅部署网络服务并不自动触发 GPLv3 的 AGPL 网络源码条款，因为本项目代码没有采用
AGPL；但实际向第三方分发二进制时 GPLv3 源码提供义务仍然适用。

## 依赖清单

Go module 清单和前端 dependency tree 位于发布包 `sbom/`。前端生产依赖许可证
清单由 pnpm lock/install 状态生成，并移除构建机路径。主要许可证族包括 MIT、
Apache-2.0、ISC、BSD、MPL-2.0、0BSD 与 BlueOak。

## 未决事项

自动清单不能替代逐包法律判断；特别是多重许可证表达式、字体/图片、间接依赖的
NOTICE 要求和商标使用应在公开分发前由发布者复核。无法确认时不得删除声明或缩小
上游权利人范围。
