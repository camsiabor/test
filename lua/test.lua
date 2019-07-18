
local id = tina.GetMicroroller().GetGateway().GetId()

NanoLocalRegister({
    Address = "qam.lua.test",
    Handler = function(msg)
        Reply(msg, 0, "test " .. id)
    end
})


