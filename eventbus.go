package main

import (
	"fmt"
	"github.com/camsiabor/qservice/core"
	"github.com/camsiabor/qservice/impl/zookeeper"
	"time"
)

var gateway core.Gateway
var overseer *core.Overseer

func initService() {

	//gateway = &memory.MGateway{}
	var zgateway = &zookeeper.ZGateway{}

	zgateway.Endpoints = []string{"127.0.0.1:12181"}
	zgateway.SessionTimeout = time.Duration(30) * time.Second

	gateway = zgateway

	overseer = &core.Overseer{}

	fmt.Println("[service] initiating")

	overseer.Init(gateway, 1024*1024)

	if err := gateway.Start(); err != nil {
		panic(err)
	}

	if err := overseer.Start(); err != nil {
		panic(err)
	}

	err := overseer.ServiceRegister("qam.echo", nil, func(message *core.Message) {
		_, _ = fmt.Printf("echo %v\n", message.Data)
		_ = message.Reply(0, message.Data)
	})

	if err != nil {
		panic(err)
	}

}
