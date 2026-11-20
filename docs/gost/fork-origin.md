# 内嵌 GOST Fork 来源

`go-gost/` 是 Agent 主模块，`go-gost/x/` 是内嵌修改版 `github.com/go-gost/x`；
`go-gost/go.mod` 用本地 `replace` 固定该 fork。许可证为 MIT，声明保存在
`LICENSES/GOST-MIT.txt`。

该目录来自 FLVX 冻结血缘，不等同于官方最新 GOST。协议或安全判断必须以当前 fork 的
registry、listener、handler、connector、dialer 和 Runtime 测试为准，不能用最新上游
文档替代。
