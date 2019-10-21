package service

import (
	"github.com/camsiabor/qservice/qtiny"
	"log"
)

func SimpleTinyGuide() *qtiny.TinyGuide {

	var guide = &qtiny.TinyGuide{}

	guide.Start = func(event qtiny.TinyGuideEvent, tiny qtiny.TinyKind, guide qtiny.TinyGuideKind, config map[string]interface{}, future *qtiny.Future, err error) {

		if err != nil {
			log.Printf("simple tiny guide start error %v", err)
			return
		}

		_ = tiny.NanoLocalRegister(qtiny.NewNano("qam.err", 0, nil, func(message *qtiny.Message) {
			_ = message.Error(500, "i am wrong. node "+tiny.GetTina().GetNodeId())
		}))
	}

	return guide

}
