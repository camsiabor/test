package eventbus

import (
	"fmt"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qservice/impl/zookeeper"
	"github.com/camsiabor/qservice/qtiny"
	"log"
	"os"
	"strings"
)

var gateway qtiny.Gateway
var overseer *qtiny.Overseer

func InitEventBus(config map[string]interface{}) {
	var clusterConfig = util.GetMap(config, true, "cluster")
	initClusterService(clusterConfig)
}

func initClusterService(config map[string]interface{}) {

	gateway = &zookeeper.ZGateway{}
	overseer = &qtiny.Overseer{}

	fmt.Println("[service] cluster initiating")

	var logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile|log.LUTC)
	gateway.SetLogger(logger)
	if err := gateway.Start(config, nil); err != nil {
		if !strings.Contains(err.Error(), "connect") {
			panic(err)
		}
	}

	overseer.SetGateway(gateway)
	if err := overseer.Start(config); err != nil {
		panic(err)
	}

	_ = overseer.NanoLocalRegister("qam.echo", 0, nil, func(message *qtiny.Message) {
		_, _ = fmt.Printf("cluster echo %v\n", message.Data)
		_ = message.Reply(0, message.Data)
	})

	_ = overseer.NanoLocalRegister("qam.ping", 0, nil, func(message *qtiny.Message) {
		_ = message.Reply(0, "pong "+overseer.GetGateway().GetId())
	})

}

func GetGateway() qtiny.Gateway {
	return gateway
}

func GetOverseer() *qtiny.Overseer {
	return overseer
}
