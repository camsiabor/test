
local id = tina.GetMicroroller().GetGateway().GetId()

NanoLocalRegister({
    Address = "qam.lua.test2",
    Handler = function(msg)
        Reply(msg, 0, "test2 " .. id)
    end
})


