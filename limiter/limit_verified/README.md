# limit verified

滑动窗口验证码限制器

## 设计说明

**滑动窗口按 target 共享, 不按 scene 隔离。** 即同一 target(手机号) 的所有场景(登录/注册/重置密码等)共享同一个滑动窗口配额。例如配置 `Quota: 30`, 则该 target 24小时内所有场景合计最多发送 30 次验证码。

验证码(code)本身按 scene 隔离存储, 不同场景的验证码互不影响。

## redis 存储格式

> key: `keyPrefix:{target}` ----> `sorted zset member`
> code key: `keyPrefix:{target}:{scene}:code` -----> `{ code, max_attempts, attempts, lasted, id }`
>
> sorted zset member: 发送时间戳 -> 唯一id
> code: code 验证码
> max_attempts: 最大允许尝试次数
> attempts: 尝试次数
> lasted: code 发送时间
> id: 唯一id
