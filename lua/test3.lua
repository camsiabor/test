
local id = tina.GetMicroroller().GetGateway().GetId()

NanoLocalRegister({
    Address = "qam.lua.test3",
    Handler = function(msg)
        Reply(msg, 0, "test3 " .. id)
    end
})





