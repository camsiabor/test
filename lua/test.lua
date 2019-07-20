
local id = tina.GetMicroroller().GetGateway().GetId()

qtiny.NanoLocalRegister({
    Address = "qam.lua.test",
    Handler = function(msg)
        qtiny.MessageReply(msg, 0, "test " .. id)
    end
})

qtiny.NanoLocalRegister({
    Address = "qam.lua.power",
    Handler = function(msg)
        qtiny.MessageReply(msg, 0, "power " .. id)
    end
})

qtiny.NanoLocalRegister({
    Address = "qam.lua.here",
    Handler = function(msg)
        qtiny.MessageReply(msg, 0, "here is the money " .. id)
    end
})


