local key = KEYS[1]                         -- 验证目标key (keyPrefix + target)
local code_key = KEYS[2]                    -- 验证码key (keyPrefix + target + ":" + scene + ":code")
local window = tonumber(ARGV[1])            -- 最大滚动窗口时间, 单位: 秒
local quota = tonumber(ARGV[2])             -- 最大滚动窗口内配额
local code_expires = tonumber(ARGV[3])      -- 验证码有效期, 单位: 秒
local code_max_attempts = tonumber(ARGV[4]) -- 验证码最大允许尝试次数
local code = ARGV[5]                        -- 验证码
local unique_id = ARGV[6]                   -- 唯一id
local tier_count = tonumber(ARGV[7])        -- 子窗口数量
local time_res = redis.call('TIME')         -- 获取redis节点当前时间.
local now = tonumber(time_res[1])           -- 当前时间戳, 单位秒

-- 检查最大窗口
redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window) -- 清除最大窗口之外的过期记录
local current_count = redis.call('ZCARD', key)            -- 统计最大窗口内的记录
if current_count >= quota then                            -- 超过最大窗口配额
    return 1                                              -- 超过配额
end

-- 检查子窗口
for i = 0, tier_count - 1 do
    local tier_window = tonumber(ARGV[8 + i * 2])
    local tier_quota = tonumber(ARGV[8 + i * 2 + 1])
    local tier_count_in_window = redis.call('ZCOUNT', key, now - tier_window, now)
    if tier_count_in_window >= tier_quota then
        return 2 -- 发送过于频繁
    end
end

-- 3. 全部通过, 记录本次操作
redis.call('ZADD', key, now, unique_id) -- 记录本次操作
redis.call('EXPIRE', key, window)       -- 设置 Key 的过期时间为最大窗口
redis.call("HSET", code_key, "code", code, "max_attempts", code_max_attempts, "attempts", 0, "lasted", now,
    "id", unique_id)
redis.call("EXPIRE", code_key, code_expires)

return 0 -- 成功
