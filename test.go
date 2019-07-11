package main

import (
	"fmt"
	"github.com/camsiabor/test/eventbus"
	"github.com/camsiabor/test/httpt"
	"time"
)

func main() {

	httpt.InitWeb()

	eventbus.InitService()

	go func() {
		for {
			fmt.Println("@heartbeat")
			time.Sleep(time.Duration(3*60) * time.Second)
		}
	}()

	<-make(chan int)
}
