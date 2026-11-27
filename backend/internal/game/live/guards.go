package live

// guards are Lua helpers every script runs before its first write, because Valkey never rolls a script back.
const guards = `local malformed = 'live: malformed live-state key'
local function hashOrNone(k)
  local t = redis.call('TYPE', k)['ok']
  return t == 'hash' or t == 'none'
end
local function integerWithRoom(v, delta)
  if not string.match(v, '^-?%d+$') then return false end
  local negative = string.sub(v, 1, 1) == '-'
  local digits = negative and string.sub(v, 2) or v
  if (#digits > 1 and string.sub(digits, 1, 1) == '0') or (negative and digits == '0') then return false end
  if #digits ~= 19 then return #digits < 19 end
  local limit
  if negative then
    limit = delta < 0 and '9223372036854775807' or '9223372036854775808'
  else
    limit = delta > 0 and '9223372036854775806' or '9223372036854775807'
  end
  return digits <= limit
end
local function counter(k, field, delta)
  if not hashOrNone(k) then return false end
  local v = redis.call('HGET', k, field)
  return v == false or integerWithRoom(v, delta)
end
`
