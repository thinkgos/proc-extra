local key = KEYS[1]                    -- 限制的Key
local locked_key = KEYS[2]             -- 锁定的Key, 用于标记当前窗口是否被锁定.
local window = tonumber(ARGV[1])       -- 有效时间窗口, 单位: 秒
local max_failures = tonumber(ARGV[2]) -- 最大允许失败次数
local unique_id = ARGV[3]              -- 唯一id, 随机串, 用于区分同一毫秒.
local is_failure = ARGV[4]             -- 是否失败操作

local time_res = redis.call('TIME')    -- 获取redis节点当前时间.
local now = tonumber(time_res[1])      -- 当前时间戳, 单位秒

-- NOTE: key和locked_key是互斥的, 绝对不会同时存在的情况.
if redis.call('EXISTS', locked_key) == 1 then -- 是否处于锁定状态
    local ttl = redis.call('TTL', locked_key)
    return { 1, now + ttl, max_failures }     -- deny, 被锁定中
end


redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window) -- 清除当前窗口之外的过期记录
local current_failures = redis.call('ZCARD', key)         -- 统计窗口内的记录

if current_failures >= max_failures then                  -- 超出限制
    local ttl = redis.call('TTL', key)
    return { 1, now + ttl, current_failures }             -- deny, 超出最大允许失败次数
else                                                      -- 未超出限制
    if is_failure == "1" then                             -- 尝试失败的验证
        redis.call('ZADD', key, now, unique_id)           -- 记录尝试失败
        redis.call('EXPIRE', key, window)                 -- 设置Key的过期时间
        return { 0, now + window, current_failures + 1 }  -- allowed, 但已记录失败
    else                                                  -- 尝试成功的验证
        redis.call('DEL', key)                            -- 并清除限制
        return { 0, now + window, 0 }                     -- allowed, 并清除限制
    end
end
