package main

import (
	"fmt"
	"github.com/camsiabor/qservice/core"
	"github.com/camsiabor/qservice/impl/memory"
)

var gateway = &memory.MemoryGateway{}
var overseer = &core.Overseer{}

func initService() {

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
