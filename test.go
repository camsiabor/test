package main

import (
	"fmt"
	"github.com/camsiabor/qservice/core"
	"github.com/camsiabor/qservice/impl"
	"time"
)

func main() {

	initWeb()

	initService()

	go func() {
		for {
			fmt.Println("heartbeat!")
			time.Sleep(time.Duration(30) * time.Second)
		}
	}()

	<-make(chan int)
}
