package main

import (
	"fmt"
	"github.com/camsiabor/qcom/qconfig"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qluatiny/qluatiny"
	"github.com/camsiabor/qservice/impl/zookeeper"
	"github.com/camsiabor/qservice/qtiny"
	"github.com/camsiabor/test/httpt"
	"github.com/camsiabor/test/service"
	"log"
	"os"
	"time"
)

func main() {

	var configPath = "config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	var config, err = qconfig.ConfigLoad(configPath, "")
	if err != nil {
		panic(err)
	}

	var tina = initTina(config)
	initTinys(tina, config)
	initHttp(config)

	<-make(chan int)
}

func initTina(config map[string]interface{}) *qtiny.Tina {
	var tina = qtiny.GetTina()
	tina.SetGateway(&zookeeper.ZGateway{})
	tina.SetMicroroller(&qtiny.Microroller{})
	var err = tina.Start(config)
	if err != nil {
		panic(err)
	}
	return tina
}

func initTinys(tina *qtiny.Tina, config map[string]interface{}) {
	var zkguide = service.ZookeeperTiny()
	var future = tina.Deploy("zookeeper", zkguide, config, 0, nil)
	go future.Run()

	var luaConfig = util.GetMap(config, true, "lua")
	var luaPath = util.GetStr(luaConfig, "../../src/github.com/camsiabor/test/lua", "path")
	var luaCPath = util.GetStr(luaConfig, luaPath, "cpath")
	var luaTinys = util.GetMap(luaConfig, true, "tinys")

	for luaMain, tinyConfigOne := range luaTinys {
		var tinyConfig = util.AsMap(tinyConfigOne, true)
		var guide = &qluatiny.LuaTinyGuide{}
		guide.Main = luaMain
		guide.LuaPath = luaPath
		guide.LuaCPath = luaCPath
		future = tina.Deploy("luatiny."+luaMain, guide, tinyConfig, 0, nil)
		future.OnFail(func(event qtiny.FutureEvent, future *qtiny.Future) qtiny.FutureCallbackReturn {
			log.Printf("deploy luatiny %v fail : %v", luaMain, future.ErrCause())
			return 0
		})

		go future.Run()
	}

}

func initHttp(config map[string]interface{}) {
	var httpConfig = util.GetMap(config, true, "http")
	httpt.InitWeb(httpConfig)
	go func() {
		for {
			fmt.Println("@heartbeat")
			time.Sleep(time.Duration(3*60) * time.Second)
		}
	}()

}

func test() {

	var queue = make(chan int, 1024)

	go func() {
		var i = 1
		for {
			queue <- i
			i++
			time.Sleep(time.Second)
		}
	}()

	for n := 1; n <= 3; n++ {
		go func(index int) {
			for {
				var i = <-queue
				fmt.Printf("consumer %v   %v\n", index, i)
			}
		}(n)
	}

	<-make(chan bool)
}
