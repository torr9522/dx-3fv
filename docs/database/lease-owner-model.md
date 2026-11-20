# Lease / Owner 模型

`IXPortResource` 描述 Provider 公网端口到节点内部 Listener 的事实；公网端口和内部端口
不能混为一谈。`IXPortLease` 按 node/internal port/protocol 分配，`LeaseKey` 只保证活动
代唯一，释放后改为 tombstone key，从而保留历史又允许再次分配。

Owner Type/ID 指向具体 IXRoute lane 或迁移 journal。`AllocateIXPorts` 在事务内校验资源、
协议映射、排除项和冲突。UI 折叠、显示名称和诊断绝不能改变 Lease。只有所属 Runtime
停止得到确认后才能请求/完成释放。
