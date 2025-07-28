-- token bucket ratelimiter script

-- `now` and refill intervals are passed as milliseconds
local now = tonumber(ARGV[1])
local globalBucketCapacity = tonumber(ARGV[2])
local globalRefillInterval = tonumber(ARGV[3])
local endpointBucketCapacity = tonumber(ARGV[4])
local endpointRefillInterval = tonumber(ARGV[5])

if now == nil or globalBucketCapacity == nil or globalRefillInterval == nil or endpointBucketCapacity == nil or endpointRefillInterval == nil or ARGV[6] == nil then
	return { err = "Invalid arguments" }
end

local globalCountKey = "global:count"
local globalLastTsKey = "global:ts"
local endpointCountKey = "endpoint:"..ARGV[6]..":count"
local endpointLastTsKey = "endpoint:"..ARGV[6]..":ts"

local fields = redis.call("HMGET", KEYS[1], globalCountKey, globalLastTsKey, endpointCountKey, endpointLastTsKey)

local expireOption = "NX"
for _, field in ipairs(fields) do
	if field then
		expireOption = "GT" -- if any of the fields exist, we will use GT to only expire if the new value is greater than the current TTL
		break
	end
end

-- calculates new count and last refill timestamp after refilling the bucket
-- for initialization it will use the max capacity as the current count
local calcNewCountAndTs = function(countField, tsField, max, refillInterval)
	local count = countField or max
	local lastRefillTs = tsField or now
	local refillAmount = math.floor((now - lastRefillTs) / refillInterval) -- how many refill intervals have passed
	count = math.min(count + refillAmount, max) -- refill the bucket, but do not exceed the max
	return count, lastRefillTs + (refillAmount * refillInterval)
end

local globalCount, globalLastTs = calcNewCountAndTs(fields[1], fields[2], globalBucketCapacity, globalRefillInterval)
local endpointCount, endpointLastTs = calcNewCountAndTs(fields[3], fields[4], endpointBucketCapacity, endpointRefillInterval)	

if globalCount <= 0 or endpointCount <= 0 then
	return 0 -- no tokens left in either bucket, block the request
end

-- consume one token from each bucket
globalCount = globalCount - 1
endpointCount = endpointCount - 1
local expiresIn = math.max((globalBucketCapacity - globalCount) * globalRefillInterval, (endpointBucketCapacity - endpointCount) * endpointRefillInterval)
redis.call("HSET", KEYS[1], globalCountKey, globalCount, globalLastTsKey, globalLastTs, endpointCountKey, endpointCount, endpointLastTsKey, endpointLastTs)
redis.call("PEXPIRE", KEYS[1], expiresIn, expireOption) -- use PEXPIRE since we are working with milliseconds

return 1
