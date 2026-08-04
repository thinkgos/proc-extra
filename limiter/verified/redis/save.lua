local key = KEYS[1]                   -- key
local value = ARGV[1]                 -- 值
local maxAttempts = tonumber(ARGV[2]) -- 最大允许尝试次数
local expires = tonumber(ARGV[3])     -- 过期时间

redis.call("HMSET", key, "value", value, "max_attempts", maxAttempts, "attempts", 0)
redis.call("EXPIRE", key, expires)
return 0 -- 成功
