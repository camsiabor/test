
local id = tina.GetMicroroller().GetGateway().GetId()
NanoLocalRegister("qam.lua.test", 0, nil, function(msg)
    Reply(msg, 0, "power overwhelming " .. id)
end)

NanoLocalRegister("qam.lua.test2", 0, nil, function(msg)
    Reply(msg, 0, "test2 " .. id)
end)



