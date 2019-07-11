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
			fmt.Println("@heartbeat")
			time.Sleep(time.Duration(3*60) * time.Second)
		}
	}()

	<-make(chan int)
}
