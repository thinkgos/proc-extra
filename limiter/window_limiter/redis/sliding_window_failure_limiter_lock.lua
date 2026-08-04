local key = KEYS[1]                   -- 限制的Key
local window = tonumber(ARGV[1])      -- 有效时间窗口, 单位: 秒
local max_failure = tonumber(ARGV[2]) -- 最大允许失败次数

local locked_key = key .. ":_locked"  -- 锁定的Key, 用于标记当前窗口是否被锁定.
local time_res = redis.call('TIME')   -- 获取redis节点当前时间.
local now = tonumber(time_res[1])     -- 当前时间戳, 单位秒

-- NOTE: key和locked_key是互斥的, 绝对不会同时存在的情况.
redis.call('DEL', key)                          -- 删除当前窗口所有旧的记录
redis.call('SET', locked_key, '', 'EX', window) -- 设置锁定的Key, 并设置过期时间为当前窗口时间.
return { 1, now + window, max_failure }         -- deny, 被强制锁定中
