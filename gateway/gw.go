package gateway

import (
	"fmt"
	"github.com/camsiabor/qservice/impl/zookeeper"
	"github.com/camsiabor/qservice/qtiny"
	"log"
	"os"
	"strings"
)

var gateway qtiny.Gateway
var microroller *qtiny.MicroRoller

func InitGateway(config map[string]interface{}) {

	gateway = &zookeeper.ZGateway{}
	microroller = &qtiny.MicroRoller{}

	fmt.Println("[service] cluster initiating")

	var logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile|log.LUTC)
	gateway.SetLogger(logger)
	if err := gateway.Start(config); err != nil {
		if !strings.Contains(err.Error(), "connect") {
			panic(err)
		}
	}

	microroller.SetGateway(gateway)
	if err := microroller.Start(config); err != nil {
		panic(err)
	}

}

func GetGateway() qtiny.Gateway {
	return gateway
}

func GetMicroroller() *qtiny.MicroRoller {
	return microroller
}
