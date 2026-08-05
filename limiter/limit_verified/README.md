# limit verified

滑动窗口验证码限制器

> redis 存储格式:  
>
> key: `keyPrefix:{target}` ----> `sorted zset member`  
> code key: `keyPrefix:{target}:_entry_:code` -----> `{ code, max_attempts, attempts, lasted, id }`  
>
> sorted zset member: 发送时间戳 -> 唯一id  
> code: code 验证码  
> max_attempts: 最大允许尝试次数  
> attempts: 尝试次数
> lasted: code 发送时间
> id: 唯一id
