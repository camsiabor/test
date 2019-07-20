
local id = tina.GetMicroroller().GetGateway().GetId()

qtiny.NanoLocalRegister({
    Address = "qam.lua.test2",
    Handler = function(msg)
        qtiny.MessageReply(msg, 0, "test 2 " .. id)
    end
})
