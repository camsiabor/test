local nodeId = tina.GetNodeId()

qtiny.NanoLocalRegister({
    Address = "qam.lua.test2",
    Handler = function(msg)
        tina.GetLogger().Println("test!")
        qmsg.Reply(msg, 0, "test " .. nodeId)
    end
})

qtiny.NanoLocalRegister({
    Address = "qam.lua.err2",
    Handler = function(msg)
        qmsg.Error(msg, 0, "testing error " .. nodeId)
    end
})

qtiny.AddCloseHandler(function()
    tina.GetLogger().Println("i am close!")
end)

