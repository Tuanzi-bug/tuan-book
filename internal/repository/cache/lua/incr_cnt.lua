-- 具体业务
local key = KEYS[1]
-- 是阅读数，点赞数还是收藏数
local cntKey = ARGV[1]
-- 增加的数量
local delta = tonumber(ARGV[2])

-- 判断是否存在
local exists = redis.call('EXISTS', key)
if exists==1 then
    redis.call('HINCRBY', key, cntKey, delta)
    return 1
else
    return 0
end