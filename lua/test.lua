local json = require "util.json"

local nodeId = tina.GetNodeId()

qtiny.NanoLocalRegister({
    Address = "qam.lua.test",
    Handler = function(msg)
        tina.GetLogger().Println("test!")

        qtiny.MsgEasy(msg)

        tina.GetLogger().Println(theM.Data)

        local data = json.decode(theM.Data)
        local str = json.encode(data)
        tina.GetLogger().Println("finally " .. str)

        qtiny.MsgReply(msg, 0, "test " .. nodeId)
    end
})

qtiny.NanoLocalRegister({
    Address = "qam.lua.err",
    Handler = function(msg)
        qtiny.MsgError(msg, 0, "testing error " .. nodeId)
    end
})

qtiny.AddCloseHandler(function ()
    tina.GetLogger().Println("i am close!")
end)

