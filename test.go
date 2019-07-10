package main

import (
	"fmt"
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
