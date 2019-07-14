package service

import (
	"fmt"
	"github.com/camsiabor/golua/lua"
	"github.com/camsiabor/golua/luar"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qservice/qtiny"
	"github.com/camsiabor/test/eventbus"
)

func InitLuaService(config map[string]interface{}) {
	var overseer = eventbus.GetOverseer(true)

	var luaPath = util.GetStr(config, "../../src/github.com/camsiabor/test/lua/", "lua", "path")
	var luaCPath = util.GetStr(config, luaPath, "lua", "cpath")

	var L, err = InitLuaState(luaPath, luaCPath)
	if err != nil {
		panic(err)
	}
	luar.Register(L, "tiny", map[string]interface{}{
		"overseer": overseer,
	})

	L.Register("ServiceRegister", func(L *lua.State) int {
		var address = L.ToString(1)
		fmt.Println(address)
		//L.SetField(lua.LUA_REGISTRYINDEX, )

		var ref = L.Ref(lua.LUA_REGISTRYINDEX)

		var err = overseer.ServiceRegister(address, nil, func(message *qtiny.Message) {
			var pointer interface{} = message
			L.RawGeti(lua.LUA_REGISTRYINDEX, ref)
			L.PushLightUserdata(&pointer)

			L.Call(2, 0)
		})

		if err == nil {
			L.PushNil()
			fmt.Println("lua service register ", address)
		} else {
			L.PushString(err.Error())
			fmt.Println("lua service register fail ", address, err.Error())
		}
		return 1

	})

	var lret, lerr = RunLuaFile(L, "test.lua", func(L *lua.State, pan interface{}) {
		if pan != nil {
			panic(pan)
		}
	})
	fmt.Println("lua return", lret)
	if lerr != nil {
		panic(lerr)
	}

	/*
		_ = overseer.ServiceRegister("qam.zk.conn", nil, func(message *qtiny.Message) {
			var _, id, endpoint, _= getParams(message)
			var _, err= zookeeper.ZooWatcherGet(id, endpoint)
			if err == nil {
				_ = message.Reply(0, id+" connected")
			} else {
				_ = message.Error(500, err.Error())
			}
		})
	*/

}
