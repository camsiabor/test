package main

import (
	"fmt"
	"github.com/camsiabor/qservice/core"
	"github.com/camsiabor/qservice/impl/memory"
	"github.com/camsiabor/qservice/impl/zookeeper"
	"time"
)

var localGateway core.Gateway
var localOverseer *core.Overseer

var clusterGateway core.Gateway
var clusterOverseer *core.Overseer

func initService() {
	initLocalService()
	initClusterService()
}

func initLocalService() {
	localGateway = &memory.MGateway{}
	localOverseer = &core.Overseer{}

	fmt.Println("[service] local initiating")

	var overconfig = map[string]interface{}{"gateway": localGateway}
	if err := localOverseer.Start(overconfig); err != nil {
		panic(err)
	}

	var memconfig = map[string]interface{}{}
	if err := localGateway.Start(memconfig); err != nil {
		panic(err)
	}

	err := localOverseer.ServiceRegister("qam.echo", nil, func(message *core.Message) {
		_, _ = fmt.Printf("local echo %v\n", message.Data)
		_ = message.Reply(0, message.Data)
	})

	if err != nil {
		panic(err)
	}
}

func initClusterService() {

	var zgateway = &zookeeper.ZGateway{}
	zgateway.Endpoints = []string{"127.0.0.1:12181"}
	zgateway.SessionTimeout = time.Duration(30) * time.Second

	clusterGateway = zgateway
	clusterOverseer = &core.Overseer{}

	fmt.Println("[service] cluster initiating")

	var overconfig = map[string]interface{}{"gateway": clusterGateway}
	if err := clusterOverseer.Start(overconfig); err != nil {
		panic(err)
	}

	var zkconfig = map[string]interface{}{
		"endpoints":       []string{"127.0.0.1:12181"},
		"session.timeout": 10,
	}
	if err := clusterGateway.Start(zkconfig); err != nil {
		panic(err)
	}

	err := clusterOverseer.ServiceRegister("qam.echo", nil, func(message *core.Message) {
		_, _ = fmt.Printf("cluster echo %v\n", message.Data)
		_ = message.Reply(0, message.Data)
	})

	if err != nil {
		panic(err)
	}
}
