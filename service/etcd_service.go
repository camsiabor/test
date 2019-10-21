package service

import (
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qservice/qtiny"
	"log"
)

func etcdGetParams(message *qtiny.Message) (request map[string]interface{}, id string, endpoint string, path string) {
	request = util.AsMap(message.Data, false)
	id = util.GetStr(request, "", "id")
	endpoint = util.GetStr(request, "127.0.0.1:2379", "endpoint")
	path = util.GetStr(request, "/", "path")
	return
}

func EtcdTinyGuide() *qtiny.TinyGuide {

	var guide = &qtiny.TinyGuide{}

	guide.Start = func(event qtiny.TinyGuideEvent, tiny qtiny.TinyKind, guide qtiny.TinyGuideKind, config map[string]interface{}, future *qtiny.Future, err error) {

		if err != nil {
			log.Printf("etcd tiny guide start error %v", err)
			return
		}

		_ = tiny.NanoLocalRegister(qtiny.NewNano("qam.etcd.conn", 0, nil, func(message *qtiny.Message) {
			var _, id, endpoint, _ = etcdGetParams(message)
			var _, err = EtcdGet(id, endpoint)
			if err == nil {
				_ = message.Reply(0, id+" connected")
			} else {
				_ = message.Error(500, err.Error())
			}
		}))

	}

	return guide
}
