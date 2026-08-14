local keyPrefix = KEYS[1]                                 -- keyPrefix
local target = KEYS[2]                                    -- target
local scene = KEYS[3]                                     -- scene
local window = tonumber(ARGV[1])                          -- 验证码滚动窗口时间, 单位: 秒
local quota = tonumber(ARGV[2])                           -- 验证码滚动窗口内配额
local resend_interval = tonumber(ARGV[3])                 -- 验证码重发间隔时间
local code_expires = tonumber(ARGV[4])                    -- 验证码有效期, 单位: 秒
local code_max_attempts = tonumber(ARGV[5])               -- 验证码最大允许尝试次数
local code = ARGV[6]                                      -- 验证码
local unique_id = ARGV[7]                                 -- 唯一id

local key = keyPrefix .. target                           -- 验证目标key
local code_key = key .. ":" .. scene .. ":code"           -- 验证码key
local time_res = redis.call('TIME')                       -- 获取redis节点当前时间.
local now = tonumber(time_res[1])                         -- 当前时间戳, 单位秒

redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window) -- 清除当前窗口之外的过期记录
local current_count = redis.call('ZCARD', key)            -- 统计窗口内的记录
if current_count >= quota then                            -- 超过配额
    return 1
else
    local lasted = redis.call("HGET", code_key, "lasted")
    if lasted then
        if tonumber(lasted) + resend_interval > now then
            return 2 -- 发送过于频繁
        end
    end

    redis.call('ZADD', key, now, unique_id) -- 记录本次操作
    redis.call('EXPIRE', key, window)       -- 设置 Key 的过期时间
    redis.call("HSET", code_key, "code", code, "max_attempts", code_max_attempts, "attempts", 0, "lasted", now,
        "id", unique_id)
    redis.call("EXPIRE", code_key, code_expires)

    return 0 -- 成功
end
