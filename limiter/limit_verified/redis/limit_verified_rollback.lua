local key = KEYS[1]       -- key
local unique_id = ARGV[1] -- 唯一id

local code_key = key .. ":_entry_:code"

redis.call("ZREM", key, unique_id)
local current_unique_id = redis.call("HGET", code_key, "id")
if current_unique_id == unique_id then
	redis.call("DEL", code_key)
end
