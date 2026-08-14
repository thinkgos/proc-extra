local key = KEYS[1]                             -- 验证目标key (keyPrefix + target)
local code_key = KEYS[2]                        -- 验证码key (keyPrefix + target + ":" + scene + ":code")
local code = ARGV[1]                            -- 验证码

if redis.call("EXISTS", code_key) == 0 then
    return 2 -- 验证码已失效
end

local vals = redis.call('HMGET', code_key, "attempts", "max_attempts", "code")
local attempts = tonumber(vals[1])
local max_attempts = tonumber(vals[2])
local current_code = vals[3]

if attempts >= max_attempts then -- 尝试次数达到最大限制
    return 2                     -- 验证码已失效
end

if current_code == code then
    redis.call("HSET", code_key, "attempts", max_attempts) -- 置失效
    return 0                                               -- 成功
else
    redis.call('HINCRBY', code_key, "attempts", 1)         -- 递增尝试次数
    return 1                                               -- 验证码错误
end
