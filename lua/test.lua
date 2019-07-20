
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

qtiny.AddCloseHandler(function ()
    tina.GetLogger().Println("i am close! 333333 333")
end)

qtiny.AddCloseHandler(function ()
    tina.GetLogger().Println("i am close! 3666 xxx")
end)

