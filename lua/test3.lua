local nodeId = tina.GetNodeId()

qtiny.NanoLocalRegister({
    Address = "qam.lua.test3",
    Handler = function(msg)
        qtiny.MessageReply(msg, 0, "test " .. nodeId)
    end
})
