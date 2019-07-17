
local id = tina.GetMicroroller().GetGateway().GetId()
NanoLocalRegister("qam.hello.world", 0, nil, function(msg)
    Reply(msg, 0, "hello world " .. id)
end)



