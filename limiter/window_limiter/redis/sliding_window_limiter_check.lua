local key = KEYS[1]                           -- 限制的Key
local locked_key = KEYS[2]                    -- 锁定的Key, 用于标记当前窗口是否被锁定.
local window = tonumber(ARGV[1])              -- 有效时间窗口, 单位: 秒
local max_limit = tonumber(ARGV[2])           -- 最大允许操作次数

local time_res = redis.call('TIME')           -- 获取redis节点当前时间.
local now = tonumber(time_res[1])             -- 当前时间戳, 单位秒

if redis.call('EXISTS', locked_key) == 1 then -- 是否处于锁定状态
    local ttl = redis.call('TTL', locked_key)
    return { 1, now + ttl, max_limit }        -- deny, 被锁定中
end

redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window) -- 清除当前窗口之外的过期记录
local current_count = redis.call('ZCARD', key)            -- 统计窗口内的记录
local ttl = redis.call('TTL', key)
if ttl < 0 then
    ttl = window
end
return { current_count + 1 <= max_limit and 0 or 1, now + ttl, current_count }
