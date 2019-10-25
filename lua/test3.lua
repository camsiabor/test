local nodeId = tina.GetNodeId()

qtiny.NanoLocalRegister({
    Address = "qam.lua.test3",
    Handler = function(msg)
        qtiny.MsgReply(msg, 0, "test " .. nodeId)
    end
})
