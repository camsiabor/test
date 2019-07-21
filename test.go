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
	"go.etcd.io/etcd/clientv3"
	"golang.org/x/net/context"
	"log"
	"os"
	"runtime"
	"time"
)

func test() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Duration(10) * time.Second,
	})
	if err != nil {
		panic(err)
	}

	put, err := client.Put(context.TODO(), "hello", "world")
	if err == nil {
		fmt.Println(put.PrevKv)
	} else {
		fmt.Println(err.Error())
	}

	for {

		get, err := client.Get(context.TODO(), "hello")

		if err == nil {
			for _, kv := range get.Kvs {
				fmt.Println(kv.String())
			}
		} else {
			fmt.Println(err.Error())
		}

		time.Sleep(time.Second * time.Duration(3))
	}
}

func main() {

	//test()
	//
	//if 1 == 1 {
	//	return
	//}

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
