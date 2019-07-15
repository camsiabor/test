package service

import (
	"fmt"
	"github.com/camsiabor/golua/lua"
	"github.com/camsiabor/golua/luar"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qservice/qtiny"
	"github.com/camsiabor/test/eventbus"
	"unsafe"
)

func InitLuaService(config map[string]interface{}) {
	var overseer = eventbus.GetOverseer()

	var luaPath = util.GetStr(config, "../../src/github.com/camsiabor/test/lua/", "lua", "path")
	var luaCPath = util.GetStr(config, luaPath, "lua", "cpath")

	var L, err = InitLuaState(luaPath, luaCPath)
	if err != nil {
		panic(err)
	}
	luar.Register(L, "tiny", map[string]interface{}{
		"overseer": overseer,
	})

	L.Register("Reply", func(L *lua.State) int {

		var ptrvalue = L.ToInteger(1)
		var ptr = unsafe.Pointer(uintptr(ptrvalue))
		var message = (*qtiny.Message)(ptr)
		var code = L.ToInteger(2)
		var reply = L.ToString(3)
		var err = message.Reply(code, reply)
		if err == nil {
			L.PushNil()
		} else {
			L.PushString(err.Error())
		}
		return 1
	})

	L.Register("NanoLocalRegister", func(L *lua.State) int {
		var address = L.ToString(1)
		var flag = L.ToInteger(2)
		// TODO options

		var ref = L.Ref(lua.LUA_REGISTRYINDEX)

		var err = overseer.NanoLocalRegister(address, qtiny.NanoFlag(flag), nil, func(message *qtiny.Message) {
			var ptrvalue = uintptr(unsafe.Pointer(message))
			L.RawGeti(lua.LUA_REGISTRYINDEX, ref)
			L.PushInteger(int64(ptrvalue))
			L.CallHandle(1, 0, func(L *lua.State, pan interface{}) {
				var err = util.AsError(pan)
				message.Error(0, err.Error())
			})
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
		_ = overseer.NanoRegister("qam.zk.conn", nil, func(message *qtiny.Message) {
			var _, id, endpoint, _= getParams(message)
			var _, err= zookeeper.ZooWatcherGet(id, endpoint)
			if err == nil {
				_ = message.Reply(0, id+" connected")
			} else {
				_ = message.Error(500, err.Error())
			}
		})
	*/

	/*
		go func() {

			var L, _= InitLuaState(luaPath, luaCPath)
			defer L.Close()

			for {
				_, _ = RunLuaFile(L, "xor.lua", func(L *lua.State, pan interface{}) {
					if pan != nil {
						fmt.Println("error", pan)
					}
				})

				time.Sleep(time.Second * time.Duration(3))
			}

		}()
	*/

}
