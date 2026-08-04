local key = KEYS[1]                  -- 限制的Key
local window = tonumber(ARGV[1])     -- 有效时间窗口, 单位: 秒
local max_limit = tonumber(ARGV[2])  -- 最大允许操作次数
local unique_id = ARGV[3]            -- 唯一id, 随机串, 用于区分同一毫秒.

local locked_key = key .. ":_locked" -- 锁定的Key, 用于标记当前窗口是否被锁定.
local time_res = redis.call('TIME')  -- 获取redis节点当前时间.
local now = tonumber(time_res[1])    -- 当前时间戳, 单位秒

-- NOTE: key和locked_key是互斥的, 绝对不会同时存在的情况.
if redis.call('EXISTS', locked_key) == 1 then -- 是否处于锁定状态
    local ttl = redis.call('TTL', locked_key)
    return { 1, now + ttl, max_limit }        -- deny, 被锁定中
end

redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window) -- 清除当前窗口之外的过期记录
local current_count = redis.call('ZCARD', key)            -- 统计窗口内的记录
if current_count < max_limit then                         -- 判断是否超出限制
    redis.call('ZADD', key, now, unique_id)               -- 记录本次操作
    redis.call('EXPIRE', key, window)                     -- 设置 Key 的过期时间
    return { 0, now + window, current_count + 1 }         -- allow
else
    local ttl = redis.call('TTL', key)
    return { 1, now + ttl, current_count } -- deny, 超出限制次数
end
