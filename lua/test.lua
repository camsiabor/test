local json = require "util.json"

local nodeId = tina.GetNodeId()

qtiny.NanoLocalRegister({
    Address = "qam.lua.test",
    Handler = function(msg)
        tina.GetLogger().Println("test!")
        --qmsg.Easy(msg)
        --local data = json.decode(theM.Data)
        qmsg.Reply(msg, 0, nodeId .. " | " .. luaunit.index)
    end
})

qtiny.NanoLocalRegister({
    Address = "qam.lua.test2",
    Handler = function(msg)
        local response, err = qmsg.Post({ Address = "qam.lua.try" })
        if err == nil then
            qmsg.Reply(msg, 0, "try called !!!!!!!!")
        else
            qmsg.Error(msg, 500, err)
        end
    end
})

local outer = 100

qtiny.NanoLocalRegister({
    Address = "qam.lua.loop",
    Handler = function(msg)
        local sum = 0
        for i = 1, 10000 do
            for j = 1, 10000 do
                sum = sum + 1
            end
        end
        outer = outer + 100
        qmsg.Reply(msg, 0, "looping! " .. outer)
    end
})

qtiny.NanoLocalRegister({
    Address = "qam.lua.sleep",
    Handler = function(msg)
        qos.Sleep(2000)
        qmsg.Reply(msg, 0, "sleep! ")
    end
})

qtiny.NanoLocalRegister({
    Address = "qam.lua.err",
    Handler = function(msg)
        qmsg.Error(msg, 0, "testing error " .. nodeId)
    end
})

qtiny.NanoLocalRegister({
    Address = "qam.lua.try",
    Handler = function(msg)
        tina.GetLogger().Println("i am trying!!!!!!")
        local reply = "trying " .. nodeId .. " | "
        local datastr = qmsg.Data(msg)
        if datastr ~= nil then
            local data = json.decode(datastr)
            reply = reply .. " " .. data["hello"]
        end
        qmsg.Reply(msg, 0, reply)
    end
})

qtiny.AddCloseHandler(function ()
    tina.GetLogger().Println("i am close!")
end)

