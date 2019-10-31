local json = require "util.json"

local nodeId = tina.GetNodeId()

qtiny.NanoLocalRegister({
    Address = "test",
    Handler = function(msg)
        tina.GetLogger().Println("test!")
        qmsg.Reply(msg, 0, "test! | " .. nodeId)
    end
})

qtiny.NanoLocalRegister({
    Address = "test.lua.call",
    Handler = function(msg)
        local response, err = qmsg.Post({ Address = "test.try" })
        if err == nil then
            qmsg.Reply(msg, 0, "try called !!!!!!!!" .. qmsg.ReplyData(response))
        else
            qmsg.Error(msg, 500, err)
        end
    end
})

local outer = 100

qtiny.NanoLocalRegister({
    Address = "test.loop",
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
    Address = "test.sleep",
    Handler = function(msg)
        qos.Sleep(2000)
        qmsg.Reply(msg, 0, "sleep! ")
    end
})

qtiny.NanoLocalRegister({
    Address = "test.err",
    Handler = function(msg)
        qmsg.Error(msg, 0, "testing error " .. nodeId)
    end
})

qtiny.NanoLocalRegister({
    Address = "test.try",
    Handler = function(msg)
        tina.GetLogger().Println("i am trying!!!!!!")
        local reply = "trying " .. nodeId .. " | "
        local datastr = qmsg.Data(msg)
        if datastr ~= nil then
            local data = json.decode(datastr)
            reply = reply .. " " .. data["hello"]
        end
        qos.Sleep(500)
        qmsg.Reply(msg, 0, reply)
    end
})

qtiny.NanoLocalRegister({
    Address = "test.redis",
    Handler = function(msg)
        qmsg.Reply()
    end
})

qtiny.NanoLocalRegister({
    Address = "test.pcall",
    Handler = function(msg)
        local ret, x = pcall(function()
            panic("i am panic")
            return 1024
        end)
        qmsg.Reply(msg, 0, "pcall! " .. ret .. " | " .. x)
    end
})



qtiny.AddCloseHandler(function ()
    tina.GetLogger().Println("i am close!")
end)

