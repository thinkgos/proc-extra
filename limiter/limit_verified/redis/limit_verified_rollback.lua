local key = KEYS[1]                             -- 验证目标key (keyPrefix + target)
local code_key = KEYS[2]                        -- 验证码key (keyPrefix + target + ":" + scene + ":code")
local unique_id = ARGV[1]                       -- 唯一id

redis.call("ZREM", key, unique_id)
local current_unique_id = redis.call("HGET", code_key, "id")
if current_unique_id == unique_id then
	redis.call("DEL", code_key)
end
