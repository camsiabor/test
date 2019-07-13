package main

import (
	"fmt"
	"github.com/camsiabor/qcom/qconfig"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/test/eventbus"
	"github.com/camsiabor/test/httpt"
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

	var eventbusConfig = util.GetMap(config, true, "eventbus")
	eventbus.InitEventBus(eventbusConfig)
	eventbus.InitZkTService()

	var httpConfig = util.GetMap(config, true, "http")
	httpt.InitWeb(httpConfig)

	go func() {
		for {
			fmt.Println("@heartbeat")
			time.Sleep(time.Duration(3*60) * time.Second)
		}
	}()

	<-make(chan int)
}
