local key = KEYS[1]   -- key
local value = ARGV[1] -- 答案

if redis.call("EXISTS", key) == 0 then
	return 1 -- 键不存在, 验证失败
end

local vals = redis.call("HMGET", key, "value", "attempts", "max_attempts")
local want_value = vals[1]
local attempts = tonumber(vals[2])
local max_attempts = tonumber(vals[3])
if want_value == value and attempts < max_attempts then
	return 0 -- 成功
else
	return 1 -- 值不相等, 验证失败
end
