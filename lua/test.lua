local nodeId = tina.GetNodeId()

qtiny.NanoLocalRegister({
    Address = "qam.lua.test",
    Handler = function(msg)
        tina.GetLogger().Println("test!")
        qtiny.MessageReply(msg, 0, "test " .. nodeId)
    end
})

qtiny.NanoLocalRegister({
    Address = "qam.lua.err",
    Handler = function(msg)
        qtiny.MessageError(msg, 0, "testing error " .. nodeId)
    end
})

qtiny.NanoLocalRegister({
    Address = "qam.lua.here",
    Handler = function(msg)
        local info = debug.getinfo(1,'S');
        qtiny.MessageReply(msg, 0, "here is the money " .. nodeId .. " " .. info.source)
    end
})

qtiny.AddCloseHandler(function ()
    tina.GetLogger().Println("i am close!")
end)

