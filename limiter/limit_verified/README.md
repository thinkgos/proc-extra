# limit verified 

验证码限制器

> redis 存储格式:  
> 
> global key:  
>   `keyPrefix:{target}` ----> `{ sendCnt -- sendCnt, codeCnt -- codeCnt }`  
> code key:  
>   `keyPrefix:{target}:_entry_:{kind}` -----> `{ code -- code, quota -- quota, err -- err, lasted -- lasted }`  
> 
> sendCnt: 发送次数  
> codeCnt: code 发送次数  
> code: code 验证码  
> quota: code 错误次数限制  
> lasted: code 发送时间


> redis 存储格式:  
> 
> global key:  
>   `keyPrefix:{target}` ----> zset?  
> code key:  
>   `keyPrefix:{target}:_entry_:code` -----> `{ code, max_attempts, attempts, lasted }`  
>
> code: code 验证码  
> max_attempts: 最大允许尝试次数  
> attempts: 尝试次数
> lasted: 发送时间
