# enumgen

简洁的枚举工具: 从 Go 枚举类型 (AST) 生成枚举辅助代码 (Value/Label/IsValid/ParseLabel 等)。

## 规则

- 仅支持字符类型和整型类型
- 字符串和整型类型的**零值**不能做为枚举有效值. 因为`ParseLabel`解析失败将返回零值.
- 枚举值或标签在该类型下必须唯一

## 相关

- 枚举注册表 (系统字典/文档等消费方的事实来源) 格式见 `github.com/thinkgos/proc/enum_spec` 目录下的 `enum-spec.md`。
