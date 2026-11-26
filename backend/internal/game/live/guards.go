package live

// guards are Lua helpers every script runs before its first write, because Valkey never rolls a script back.
const guards = `local malformed = 'live: malformed live-state key'
local function hashOrNone(k)
  local t = redis.call('TYPE', k)['ok']
  return t == 'hash' or t == 'none'
end
local function counter(k, field)
  if not hashOrNone(k) then return false end
  local v = redis.call('HGET', k, field)
  return v == false or tonumber(v) ~= nil
end
`
