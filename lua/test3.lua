

require("util")

local id = tina.GetMicroroller().GetGateway().GetId()
NanoLocalRegister("qam.black", 0, nil, function(msg)
    Reply(msg, 0, "black sheep wall " .. power(id))
end)

