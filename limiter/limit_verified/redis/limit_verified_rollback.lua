local keyPrefix = KEYS[1]                       -- keyPrefix
local target = KEYS[2]                          -- target
local scene = KEYS[3]                           -- scene
local unique_id = ARGV[1]                       -- 唯一id

local key = keyPrefix .. target                 -- 验证目标key
local code_key = key .. ":" .. scene .. ":code" -- 验证码key

redis.call("ZREM", key, unique_id)
local current_unique_id = redis.call("HGET", code_key, "id")
if current_unique_id == unique_id then
	redis.call("DEL", code_key)
end
