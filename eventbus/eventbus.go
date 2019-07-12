package eventbus

import (
	"fmt"
	"github.com/camsiabor/qservice/impl/memory"
	"github.com/camsiabor/qservice/impl/zookeeper"
	"github.com/camsiabor/qservice/qtiny"
	"strings"
)

var localGateway qtiny.Gateway
var localOverseer *qtiny.Overseer

var clusterGateway qtiny.Gateway
var clusterOverseer *qtiny.Overseer

func InitEventBus() {
	initLocalService()
	initClusterService()
}

func initLocalService() {

	localGateway = &memory.MGateway{}
	localOverseer = &qtiny.Overseer{}

	fmt.Println("[service] local initiating")

	var memconfig = map[string]interface{}{}
	if err := localGateway.Start(memconfig); err != nil {
		panic(err)
	}

	var overconfig = map[string]interface{}{"gateway": localGateway}
	if err := localOverseer.Start(overconfig); err != nil {
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

func initClusterService() {

	clusterGateway = &zookeeper.ZGateway{}
	clusterOverseer = &qtiny.Overseer{}

	fmt.Println("[service] cluster initiating")

	var zkconfig = map[string]interface{}{
		"id":              "cluster",
		"endpoints":       []string{"127.0.0.1:12181"},
		"session.timeout": 10,
	}
	if err := clusterGateway.Start(zkconfig); err != nil {
		if !strings.Contains(err.Error(), "connect") {
			panic(err)
		}
	}

	var overconfig = map[string]interface{}{"gateway": clusterGateway}
	if err := clusterOverseer.Start(overconfig); err != nil {
		panic(err)
	}

	err := clusterOverseer.ServiceRegister("qam.echo", nil, func(message *qtiny.Message) {
		_, _ = fmt.Printf("cluster echo %v\n", message.Data)
		_ = message.Reply(0, message.Data)
	})

	if err != nil {
		panic(err)
	}
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
