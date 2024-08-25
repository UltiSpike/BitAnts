-- 定义局部变量
local rate = tonumber(ARGV[1]) -- 令牌生成速率
local cap = tonumber(ARGV[2]) -- 令牌桶的最大容量
local now = tonumber(ARGV[3]) -- 当前时间戳
local requested = tonumber(ARGV[4]) -- 请求的令牌数

-- 计算填充时间，即填满令牌桶所需的时间
local fill_time = cap / rate
-- 计算TTL（Time to Live），设置为填充时间的两倍
local ttl = math.floor(fill_time * 2)

-- 获取当前令牌数，如果不存在则初始化为桶的最大容量
local last_tokens = tonumber(redis.call("get", KEYS[1]))
if last_tokens == nil then
    last_tokens = cap
end

-- 获取上次刷新时间，如果不存在则设置为0
local last_refresh = tonumber(redis.call("get", KEYS[2]))
if last_refresh == nil then
    last_refresh = 0
end

-- 计算时间差，即当前时间与上次刷新时间的差值
local delta = math.max(0, now - last_refresh)

-- 计算在时间差内应该填充的令牌数，不超过桶的最大容量
local filled_tokens = math.min(cap, last_tokens + (delta * rate))

-- 判断请求的令牌数是否小于或等于当前填充的令牌数
local allowed = filled_tokens >= requested

-- 计算剩余的令牌数，如果请求被允许，则减去请求的令牌数
local new_tokens = filled_tokens
if allowed then
    new_tokens = filled_tokens - requested
end

-- 更新令牌数，使用setx命令设置键的值并设置生存时间ttl
redis.call("setx", KEYS[1], ttl, new_tokens)

-- 更新上次刷新时间，同样使用setx命令
redis.call("setx", KEYS[2], ttl, now)

-- 返回是否允许请求
return allowed