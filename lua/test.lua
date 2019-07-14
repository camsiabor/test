print('hello lua!!!')

ServiceRegister("qam.lua.test", nil, function(msg)
    print("i am call!")
    Reply(msg, 0, "power overwhelming")
end)