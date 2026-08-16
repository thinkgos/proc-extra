local key = KEYS[1]                       -- 存储令牌的key
local rate = tonumber(ARGV[1])            -- 令牌生成速率 (即每秒/每毫秒生成多少个令牌)
local capacity = tonumber(ARGV[2])        -- 桶的最大容量 (允许的最大突发流量/蓄水上限)
local now = tonumber(ARGV[3])             -- 当前时间戳 (单位需与 rate 的时间单位保持一致，通常为秒)
local requested_token = tonumber(ARGV[4]) -- 本次请求消耗的令牌数 (通常为 1)

local fill_time = capacity / rate         -- 计算把空桶彻底填满需要的时间 (单位: 秒/毫秒)
local ttl = math.floor(fill_time * 2)     -- 计算 Redis Key 的过期时间 TTL (填满时间的 2 倍)，防止无用 Key 长期占用内存

local last_tokens = capacity              -- 获取上次剩余的令牌数；若不存在（首次访问），则默认初始化为满桶
local last_refreshed = 0                  -- 获取上次刷新的时间戳；若不存在，则默认初始化为 0
local last_value = redis.call("GET", key)
if type(last_value) == "string" then
    local pos = string.find(last_value, ":", 1, true) -- 解析 {令牌数}:{当前时间戳}
    if pos then
        last_tokens = tonumber(string.sub(last_value, 1, pos - 1)) or capacity
        last_refreshed = tonumber(string.sub(last_value, pos + 1)) or 0
    end
end

local delta = math.max(0, now - last_refreshed)                        -- 计算自上次请求以来过去的时间差 delta (防止时间倒流，取 >= 0)
local filled_tokens = math.min(capacity, last_tokens + (delta * rate)) -- 根据时间差计算新生成的令牌数，并加上上次剩余令牌，取 min 不超过桶的最大容量
local allowed = filled_tokens >= requested_token                       -- 判断当前桶里的总令牌数是否足够本次消耗
local new_tokens = filled_tokens
if allowed then                                                        -- 如果令牌足够，扣减本次请求需要的令牌数；若不够，令牌数维持原样
    new_tokens = filled_tokens - requested_token
end

local new_value = new_tokens .. ":" .. now -- {令牌数}:{当前时间戳} 拼接为新的 value
redis.call("SETEX", key, ttl, new_value)   -- 将更新后的令牌数和当前时间戳存回 Redis，并重新设置 TTL 自动过期

return allowed
