package main

import (
	"fmt"
	"github.com/camsiabor/qcom/qconfig"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qservice/impl/gateway/zookeeper"
	"github.com/camsiabor/qservice/impl/tiny/ankotiny"
	"github.com/camsiabor/qservice/impl/tiny/luatiny"
	"github.com/camsiabor/qservice/qtiny"
	"github.com/camsiabor/test/httpt"
	"github.com/camsiabor/test/service"
	"log"
	"os"
	"runtime"
	"time"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

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

	for k, v := range luaConfig {
		var configPath = util.AsStr(v, "")
		var guide = luatiny.NewLuaTinyGuide(k, configPath)
		future = tina.Deploy("luatiny."+k+"."+configPath, guide, nil, 0, nil)
		future.OnFail(func(event qtiny.FutureEvent, future *qtiny.Future) qtiny.FutureCallbackReturn {
			log.Printf("deploy lua tiny %v fail : %v %v", k, configPath, future.ErrCause())
			return 0
		})
		go future.Run()
	}

	var ankoConfig = util.GetMap(config, true, "anko")
	var ankoPath = util.GetStr(ankoConfig, "../../src/github.com/camsiabor/test/anko", "path")
	var ankoTinys = util.GetMap(ankoConfig, true, "tinys")
	for ankoMain, ankoConfigOne := range ankoTinys {
		var tinyConfig = util.AsMap(ankoConfigOne, true)
		var guide = ankotiny.NewAnkoTinyGuide(ankoPath, ankoMain)
		future = tina.Deploy("ankotiny."+ankoMain, guide, tinyConfig, 0, nil)
		future.OnFail(func(event qtiny.FutureEvent, future *qtiny.Future) qtiny.FutureCallbackReturn {
			log.Printf("deploy anko tiny %v fail : %v", ankoMain, future.ErrCause())
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
