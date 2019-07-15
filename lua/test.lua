print('hello lua!!!')

local id = tiny.overseer.GetGateway().GetId()
print("lua id", id)
NanoLocalRegister("qam.lua.test", 0, nil, function(msg)
    print("i am call!")
    Reply(msg, 0, "power overwhelming " .. id)
end)

NanoLocalRegister("qam.lua.test2", 0, nil, function(msg)
    print("i am call! 2")
    Reply(msg, 0, "test2 " .. id)
end)



