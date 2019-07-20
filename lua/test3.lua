
local id = tina.GetMicroroller().GetGateway().GetId()

qtiny.NanoLocalRegister({
    Address = "qam.lua.test3",
    Handler = function(msg)
        qtiny.MessageReply(msg, 0, "test 3 " .. id)
    end
})
