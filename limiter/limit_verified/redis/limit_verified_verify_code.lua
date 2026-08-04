local key = KEYS[1]                     -- key
local code = ARGV[1]                    -- 验证码

local code_key = key .. ":_entry_:code" -- 验证码key

if redis.call("EXISTS", code_key) == 0 then
    return 2 -- 验证码已失效
end

local vals = redis.call('HMGET', code_key, "attempts", "max_attempts", "code")
local attempts = tonumber(vals[1])
local max_attempts = tonumber(vals[2])
local current_code = vals[3]
if current_code == code then
    redis.call("DEL", code_key) -- 删除 code key
    return 0                    -- 成功
else
    if attempts + 1 >= max_attempts then
        redis.call("DEL", code_key) -- 删除 code key
    else
        redis.call('HINCRBY', code_key, "attempts", 1)
    end
    return 1 -- 验证码错误
end
