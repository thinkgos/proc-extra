local key = KEYS[1]                  -- refresh token store key
local grace_key = KEYS[2]            -- grace store key
local request_rt_id = ARGV[1]        -- request refresh token id
local new_rt_id = ARGV[2]            -- new refresh token id
local new_rt = ARGV[3]               -- new refresh token
local new_at = ARGV[4]               -- new access token
local new_at_exp = tonumber(ARGV[5]) -- new access token expires
local new_rt_ttl = tonumber(ARGV[6]) -- new refresh token ttl
local grace_ttl = tonumber(ARGV[7])  -- grace ttl

local current_rt_id = redis.call('GET', key)
if not current_rt_id then -- 不存在, 则拒绝
    return { 1, "", "" }
end

local vals = redis.call('HMGET', grace_key, "rt_id", "at", "at_exp", "rt")

local grace_rt_id = vals[1]
local grace_at = vals[2]
local grace_at_exp = tonumber(vals[3])
local grace_rt = vals[4]
if grace_rt_id and grace_rt_id == request_rt_id then -- 在宽限期缓存期内, request_rt_id与grace_rt_id一致, 返回缓存
    return { 0, grace_at, grace_at_exp, grace_rt }
end

if current_rt_id ~= request_rt_id then -- request_rt_id与当前current_rt_id不一致, request_rt_id已被吊销, 拒绝
    return { 1, "", 0, "" }
end

-- 执行轮转逻辑
-- 缓存宽限期 rt_id, at, rt, 并设置过期时间, 过期时间不得大于rt_id的过期时间
redis.call(
    'HSET', grace_key,
    "rt_id", current_rt_id,
    "at", new_at,
    "at_exp", new_at_exp,
    "rt", new_rt
)
local cur_rt_id_ttl = redis.call("TTL", key)
if cur_rt_id_ttl > 0 then
    grace_ttl = math.min(cur_rt_id_ttl, grace_ttl)
end
redis.call("EXPIRE", grace_key, grace_ttl)
redis.call("SET", key, new_rt_id, "EX", new_rt_ttl)
return { 0, new_at, new_at_exp, new_rt }
