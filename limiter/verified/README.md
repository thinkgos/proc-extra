# verified

`verified` 包括以下两种验证器:

- `captcha`: 用于图形验证器或问答式验证器
- `reflux`: 用于签发验证.

验证方式:

成功立即失效, 失败超过最大尝试次数后失效
  > redis 存储格式:
  > captcha: `keyPrefix:{kind}:{id}` -----> `{ value -- answer, maxAttempts -- maxAttempts, attempts -- attempts }`  
  > reflux:  `keyPrefix:{kind}:{key}` -----> `{ value -- unique, maxAttempts -- maxAttempts, attempts -- attempts }`
  > > value: 验证值  
  > > maxAttempts: 最大允许尝试次数
  > > attempts: 已尝试次数
