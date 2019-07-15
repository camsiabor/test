

print('=====================')

local function encode(str, key)
    local result = {}
    local key_index = 1
    local res_index = 1
    local str_len = #str
    local key_len = #key
    for i = 1, str_len do
        local byte_str = string.byte(str, i)
        local byte_cip = string.byte(key, key_index)
        result[res_index] = byte_str ~ byte_cip
        result[res_index + 1] = ' '
        res_index = res_index + 2
        key_index = key_index + 1
        if key_index > key_len then
            key_index = 1
        end
    end
    return table.concat(result)
end

local function decode(str, key)
    local result = {}
    local key_len = #key
    local key_index = 1
    local res_index = 1
    for token in string.gmatch(str, "[^%s]+") do
        local byte_str = tonumber(token)
        local byte_cip = string.byte(key, key_index)
        local char = string.char(byte_str ~ byte_cip)
        result[res_index] = char
        res_index = res_index + 1
        key_index = key_index + 1
        if key_index > key_len then
            key_index = 1
        end
    end
    return table.concat(result)
end

local result = encode('power', 'over')
print(result)

result = decode(result, 'over')
print(result)

local numstr = '123456789.123456789'

local num = tonumber(numstr)

print(string.format("%.14f", num))

--print(3)



