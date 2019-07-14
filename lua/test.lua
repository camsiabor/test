print('hello lua!!!')

local id = tiny.overseer.GetGateway().GetId()
print("lua id", id)
ServiceRegister("qam.lua.test", nil, function(msg)
    print("i am call!")
    Reply(msg, 0, "power overwhelming " .. id)
end)

ServiceRegister("qam.lua.test2", nil, function(msg)
    print("i am call! 2")
    Reply(msg, 0, "test2 " .. id)
end)