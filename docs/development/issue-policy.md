# Issue / Commit / Tag 规则

Issue 文件必须包含版本、根因、范围、测试、风险和回滚。一个问题对应一个 Commit 和
一个 Annotated Tag；文档、安装器缺陷和 Runtime 缺陷不可混合。Tag 不覆盖、不删除。

状态为 PASS 需要对应层的真实证据；外部依赖、长期测试或根因未知应标记 PARTIAL/
Known Issue。禁止通过增加重试、放宽断言、手工改数据库或隐藏日志宣布修复。
