local nodeId = tina.GetNodeId()

qtiny.NanoLocalRegister({
    Address = "qam.lua.test2",
    Handler = function(msg)
        tina.GetLogger().Println("test!")
        qtiny.MessageReply(msg, 0, "test " .. nodeId)
    end
})

qtiny.NanoLocalRegister({
    Address = "qam.lua.err2",
    Handler = function(msg)
        qtiny.MessageError(msg, 0, "testing error " .. nodeId)
    end
})

qtiny.AddCloseHandler(function()
    tina.GetLogger().Println("i am close!")
end)

