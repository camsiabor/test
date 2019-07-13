package eventbus

import (
	"fmt"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qservice/impl/memory"
	"github.com/camsiabor/qservice/impl/zookeeper"
	"github.com/camsiabor/qservice/qtiny"
	"strings"
)

var localGateway qtiny.Gateway
var localOverseer *qtiny.Overseer

var clusterGateway qtiny.Gateway
var clusterOverseer *qtiny.Overseer

func InitEventBus(config map[string]interface{}) {
	var localConfig = util.GetMap(config, true, "local")
	var clusterConfig = util.GetMap(config, true, "cluster")
	initLocalService(localConfig)
	initClusterService(clusterConfig)
}

func initLocalService(config map[string]interface{}) {

	localGateway = &memory.MGateway{}
	localOverseer = &qtiny.Overseer{}

	fmt.Println("[service] local initiating")

	if err := localGateway.Start(config); err != nil {
		panic(err)
	}

	localOverseer.SetGateway(localGateway)
	if err := localOverseer.Start(config); err != nil {
		panic(err)
	}

	err := localOverseer.ServiceRegister("qam.echo", nil, func(message *qtiny.Message) {
		_, _ = fmt.Printf("local echo %v\n", message.Data)
		_ = message.Reply(0, message.Data)
	})

	if err != nil {
		panic(err)
	}
}

func initClusterService(config map[string]interface{}) {

	clusterGateway = &zookeeper.ZGateway{}
	clusterOverseer = &qtiny.Overseer{}

	fmt.Println("[service] cluster initiating")

	if err := clusterGateway.Start(config); err != nil {
		if !strings.Contains(err.Error(), "connect") {
			panic(err)
		}
	}

	clusterOverseer.SetGateway(clusterGateway)
	if err := clusterOverseer.Start(config); err != nil {
		panic(err)
	}

	_ = clusterOverseer.ServiceRegister("qam.echo", nil, func(message *qtiny.Message) {
		_, _ = fmt.Printf("cluster echo %v\n", message.Data)
		_ = message.Reply(0, message.Data)
	})

	_ = clusterOverseer.ServiceRegister("qam.ping", nil, func(message *qtiny.Message) {
		_ = message.Reply(0, "pong "+clusterOverseer.GetGateway().GetId())
	})

}

func GetGateway(local bool) qtiny.Gateway {
	if local {
		return localGateway
	}
	return clusterGateway
}

func GetOverseer(local bool) *qtiny.Overseer {
	if local {
		return localOverseer
	}
	return clusterOverseer
}
