local key = KEYS[1]   -- key
local value = ARGV[1] -- 答案

if redis.call("EXISTS", key) == 0 then
	return 1 -- 键不存在, 验证失败
end

local want_value = redis.call("HGET", key, "value")
if want_value == value then
	redis.call("DEL", key)
	return 0 -- 成功
else
	local max_attempts = tonumber(redis.call("HGET", key, "max_attempts"))
	local attempts = redis.call("HINCRBY", key, "attempts", 1)
	if attempts >= max_attempts then
		redis.call("DEL", key)
	end
	return 1 -- 值不相等, 验证失败
end
