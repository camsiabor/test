package main

import (
	"fmt"
	"github.com/camsiabor/qcom/qconfig"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qservice/impl/zookeeper"
	"github.com/camsiabor/qservice/qtiny"
	"github.com/camsiabor/test/httpt"
	"github.com/camsiabor/test/service"
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

	var tina = qtiny.GetTina()
	tina.SetGateway(&zookeeper.ZGateway{})
	tina.SetMicroroller(&qtiny.Microroller{})
	tina.Start(config)

	var guide = &qtiny.TinyGuide{
		Start: func(tiny qtiny.TinyKind, future qtiny.Future) error {
			tiny.Register("", 0, nil, func(message *qtiny.Message) {

			})

			return nil
		},
	}

	var err = tina.Deploy(guide, nil, 0, nil)
	if err != nil {
		panic(err)
	}

	service.InitZkTService()
	service.InitLuaService(config)

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
