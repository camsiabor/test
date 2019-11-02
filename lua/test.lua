local json = require "util.json"

local logger = qam.tina.GetLogger()
local nodeId = qam.tina.GetNodeId()

qtiny.NanoLocalRegister({
    Address = "test",
    Handler = function(msg)
        logger.Println("test!")
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
        logger.Println("i am trying!!!!!!")
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

        local ok, a, b, c = xpcall(function()
            error("i am error")
        end, function(err)
            return "power", "over", "whelming " .. err
        end)

        --[[
        local ok, a, b, c = pcall(function()
            return "power", "over", "whelming"
        end)
        ]]--

        if ok then
            qmsg.Reply(msg, 0, "pcall! " .. a .. " | " .. b .. " | " .. c)
        else
            qmsg.Error(msg, 500, "pcall! " .. a)
        end
    end
})



qtiny.AddCloseHandler(function ()
    logger.Println("i am close!")
end)

