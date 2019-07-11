package main

import (
	"fmt"
	"github.com/camsiabor/test/eventbus"
	"github.com/camsiabor/test/httpt"
	"github.com/camsiabor/test/zkt"
	"time"
)

func main() {

	eventbus.InitEventBus()

	zkt.InitZkTService()

	httpt.InitWeb(":5555")

	go func() {
		for {
			fmt.Println("@heartbeat")
			time.Sleep(time.Duration(3*60) * time.Second)
		}
	}()

	<-make(chan int)
}
