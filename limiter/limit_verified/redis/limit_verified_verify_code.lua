local keyPrefix = KEYS[1]                       -- keyPrefix
local target = KEYS[2]                          -- target
local scene = KEYS[3]                           -- scene
local code = ARGV[1]                            -- 验证码

local key = keyPrefix .. target                 -- 验证目标key
local code_key = key .. ":" .. scene .. ":code" -- 验证码key

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
